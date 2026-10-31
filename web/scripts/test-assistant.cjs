// Executes the component's real handlers with simulated Web Speech APIs.
const { readFileSync } = require('node:fs');
const { resolve } = require('node:path');
const vm = require('node:vm');
const assert = require('node:assert/strict');
const { test } = require('node:test');
const ts = require('typescript');
const source = ts.createSourceFile('assistant.tsx', readFileSync(resolve(__dirname, '../src/features/dashboard/components/JarvisChatDrawer.tsx'), 'utf8'), ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX);
const wanted = new Set(['handleSend', 'handleMicClick', 'speakText']);
const handlers = [];
function visit(node) {
  if (ts.isVariableDeclaration(node) && wanted.has(node.name.getText(source))) handlers.push(`const ${node.getText(source)};`);
  ts.forEachChild(node, visit);
}
visit(source);
assert.equal(handlers.length, 3);
const code = ts.transpileModule(handlers.join('\n') + '\nglobalThis.handlers = {handleSend, handleMicClick, speakText};', { compilerOptions: { target: ts.ScriptTarget.ES2020 } }).outputText;
function harness(persona, fail = false) {
  const calls = [], messages = [], recognitions = [], spoken = [];
  class Recognition {
    constructor() { recognitions.push(this); }
    start() { this.onstart?.(); }
    stop() {}
    abort() { this.onend?.(); }
  }
  const state = {
    persona, voiceEnabled: true, rate: 1, pitch: 1, selectedVoiceURI: '', availableVoices: [],
    activeRecognitionRef: {current: null}, isSpeakingRef: {current: false},
    lastHandledTextRef: {current: {text: '', time: 0}},
    window: {SpeechRecognition: Recognition, speechSynthesis: {cancel() {}, speak(u) {spoken.push(u);}}},
    SpeechSynthesisUtterance: class {constructor(text) {this.text = text;}},
    setMessages(fn) { messages.splice(0, messages.length, ...fn(messages)); },
    setIsThinking() {}, setIsListening() {}, setInputVal() {},
    async onExecuteCommand(text) {calls.push(text); if(fail) throw Error('API indisponível'); return 'Pipeline iniciado';},
    processAssistantResponse() {throw Error('Unexpected local fallback');},
    setTimeout, console,
  };
  vm.createContext(state); vm.runInContext(code, state);
  return {state, calls, messages, recognitions, spoken, ...state.handlers};
}
const flush = () => new Promise(resolve => setImmediate(resolve));
for (const persona of ['jarvis', 'jake']) {
  test(`${persona}: one command per speech event and no old recognition echo`, async () => {
    const h = harness(persona);
    h.handleMicClick(); const first = h.recognitions[0];
    const event = {results: [[{transcript: `${persona}, construir landing`}]]};
    first.onresult(event); first.onresult(event); await flush();
    assert.equal(h.calls.length, 1);
    assert.equal(h.messages.length, 2);
    assert.equal(h.spoken.length, 1);
    // A distinct late result must also be discarded, beyond text deduplication.
    first.onresult({results: [[{transcript: 'Pipeline iniciado'}]]}); await flush();
    assert.equal(h.calls.length, 1);
    h.handleMicClick(); const second = h.recognitions[1];
    first.onend();
    assert.equal(h.state.activeRecognitionRef.current, second);
    second.onresult({results: [[{transcript: 'status'}]]}); await flush();
    assert.equal(h.calls.length, 2);
  });
  test(`${persona}: API failure is reported once`, async () => {
    const h = harness(persona, true);
    await h.handleSend('construir projeto');
    assert.equal(h.calls.length, 1);
    assert.equal(h.messages.at(-1).text, 'Aviso operacional: API indisponível');
    assert.equal(h.spoken.length, 1);
  });
}

const pageSource = ts.createSourceFile('page.tsx', readFileSync(resolve(__dirname, '../src/app/page.tsx'), 'utf8'), ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX);
let sessionGuard, sessionEffect;
function findSessionCode(node) {
  if (ts.isFunctionDeclaration(node) && node.name?.text === 'hasMfaSession') sessionGuard = node.getText(pageSource);
  if (ts.isCallExpression(node) && node.expression.getText(pageSource) === 'useEffect' && node.arguments[0]?.getText(pageSource).includes('const forceDemo')) sessionEffect = node.arguments[0].getText(pageSource);
  ts.forEachChild(node, findSessionCode);
}
findSessionCode(pageSource);
assert.ok(sessionGuard && sessionEffect);
const sessionCode = ts.transpileModule(`${sessionGuard}\nglobalThis.boot = ${sessionEffect};`, {compilerOptions:{target:ts.ScriptTarget.ES2020}}).outputText;
function fakeSession(aal) {
  return {access_token:`header.${Buffer.from(JSON.stringify({aal})).toString('base64url')}.signature`,user:{email:'test@example.invalid'}};
}
async function sessionHarness(session, fail=false, demo=false) {
  let current=null, authCallback, isDemo=false;
  const context = {
    window:{location:{search:demo?'?demo=true':''}}, URLSearchParams, atob,
    localStorage:{getItem(){return null;}}, sessionStorage:{removeItem(){}},
    setIsDemoMode(v){isDemo=v;},setLoadingAuth(){},setSession(v){current=v;},
    supabase:{auth:{getSession:async()=>{if(fail) throw Error('offline');return {data:{session}};},onAuthStateChange(fn){authCallback=fn;return {data:{subscription:{unsubscribe(){}}}};}}},
  };
  vm.createContext(context);vm.runInContext(sessionCode,context);context.boot();await flush();
  return {get session(){return current;},get demo(){return isDemo;},authCallback};
}
test('No authenticated session or auth failure never creates an owner session',async()=>{
  for(const fail of [false,true]){const h=await sessionHarness(null,fail);assert.equal(h.session,null);assert.equal(h.demo,false);}
});
test('AAL1 cannot open owner panel; AAL2 can, and signout clears the panel',async()=>{
  assert.equal((await sessionHarness(fakeSession('aal1'))).session,null);
  const session=fakeSession('aal2');const h=await sessionHarness(session);assert.equal(h.session,session);
  h.authCallback('SIGNED_OUT',null);assert.equal(h.session,null);
});
test('Explicit demonstration stays labeled as demonstration',async()=>{
  const h=await sessionHarness(null,false,true);assert.equal(h.demo,true);assert.equal(h.session.user.id,'owner-demo');
});
