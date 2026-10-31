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

const loginSource = ts.createSourceFile('LoginForm.tsx', readFileSync(resolve(__dirname, '../src/features/auth/components/LoginForm.tsx'), 'utf8'), ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX);
const loginHandlers=[];
function findLoginHandlers(node){
 if(ts.isFunctionDeclaration(node)&&['handlePasswordSubmit','handleTOTPSubmit'].includes(node.name?.text))loginHandlers.push(node.getText(loginSource));
 ts.forEachChild(node,findLoginHandlers);
}
findLoginHandlers(loginSource);assert.equal(loginHandlers.length,2);
const loginCode=ts.transpileModule(loginHandlers.join('\n')+'\nglobalThis.loginHandlers={handlePasswordSubmit,handleTOTPSubmit};',{compilerOptions:{target:ts.ScriptTarget.ES2020}}).outputText;
function loginHarness({verified=false,listError=false,enrollError=false,verifyError=false}={}){
 const calls=[];
 const state={
  Error, email:'fixture@example.invalid',password:'fixture-password',factorId:null,totpCode:'123456',enrollment:null,step:'credentials',errorMsg:null,success:0,
  setErrorMsg(v){state.errorMsg=v;},setLoading(){},setPassword(v){state.password=v;},
  setFactorId(v){state.factorId=v;},setEnrollment(v){state.enrollment=v;},setStep(v){state.step=v;},setTotpCode(v){state.totpCode=v;},onSuccess(){state.success++;},
  supabase:{auth:{
   signInWithPassword:async()=>({error:null}),
   mfa:{
    listFactors:async()=>({data:{totp:verified?[{id:'verified-factor',status:'verified'}]:[]},error:listError?Error('list failed'):null}),
    enroll:async()=>{calls.push('enroll');return {data:{id:'new-factor',totp:{qr_code:'data:image/svg+xml;utf-8,<svg/>',secret:'FAKE-TEST-SECRET'}},error:enrollError?Error('enrollment failed'):null};},
    challenge:async({factorId})=>{calls.push(['challenge',factorId]);return {data:{id:'challenge'},error:null};},
    verify:async(args)=>{calls.push(['verify',args.factorId,args.code]);return {error:verifyError?Error('invalid code'):null};},
   },
  }},
 };
 vm.createContext(state);vm.runInContext(loginCode,state);return {state,calls,...state.loginHandlers};
}
const submit={preventDefault(){}};
test('First login enrolls TOTP and waits for confirmation before allowing access',async()=>{
 const h=loginHarness();await h.handlePasswordSubmit(submit);
 assert.equal(h.state.step,'enroll');assert.equal(h.state.factorId,'new-factor');assert.ok(h.state.enrollment.qrCode);assert.equal(h.state.success,0);assert.equal(h.state.password,'');
 await h.handleTOTPSubmit(submit);assert.equal(h.state.success,1);assert.equal(h.state.enrollment,null);assert.equal(h.state.totpCode,'');
 assert.deepEqual(h.calls.at(-1),['verify','new-factor','123456']);
});
test('Existing verified TOTP is challenged without enrolling another factor',async()=>{
 const h=loginHarness({verified:true});await h.handlePasswordSubmit(submit);assert.equal(h.state.step,'totp');assert.equal(h.state.success,0);assert.equal(h.calls.length,0);
 await h.handleTOTPSubmit(submit);assert.equal(h.state.success,1);assert.deepEqual(h.calls[0],['challenge','verified-factor']);
});
test('Factor listing and enrollment failures cannot open the panel',async()=>{
 for(const options of [{listError:true},{enrollError:true}]){const h=loginHarness(options);await h.handlePasswordSubmit(submit);assert.equal(h.state.success,0);assert.equal(h.state.step,'credentials');assert.ok(h.state.errorMsg);}
});
test('Invalid verification preserves enrollment for retry and never allows access',async()=>{
 const h=loginHarness({verifyError:true});await h.handlePasswordSubmit(submit);await h.handleTOTPSubmit(submit);
 assert.equal(h.state.success,0);assert.ok(h.state.enrollment);assert.equal(h.state.errorMsg,'invalid code');
});
test('Malformed TOTP codes are rejected before calling the provider',async()=>{
 const h=loginHarness({verified:true});await h.handlePasswordSubmit(submit);h.state.totpCode='123';await h.handleTOTPSubmit(submit);assert.equal(h.calls.length,0);assert.equal(h.state.success,0);assert.ok(h.state.errorMsg);
});
