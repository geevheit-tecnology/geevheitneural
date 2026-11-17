// Runs the actual panel's rendered controls and handlers with local API/hook doubles.
import { readFileSync } from 'node:fs';
import vm from 'node:vm';
import assert from 'node:assert/strict';
import { test } from 'node:test';
import ts from 'typescript';
import * as React from 'react';
import * as jsxRuntime from 'react/jsx-runtime';
import { renderToStaticMarkup } from 'react-dom/server';
const code = ts.transpileModule(readFileSync(new URL('../src/features/dashboard/components/FactoryToolsPanel.tsx', import.meta.url), 'utf8'), {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020, jsx: ts.JsxEmit.ReactJSX },
}).outputText;

function panel({ confirmed = false, brief = '', demo = false, configured = true, error = false, clipboardFails = false } = {}) {
  let stateIndex = 0;
  const calls = [], copied = [], messages = [];
  let queryOptions, mutationOptions;
  const componentModule = { exports: {} };
  const context = {
    module: componentModule, exports: componentModule.exports,
    navigator: { clipboard: { async writeText(text) { if (clipboardFails) throw Error('denied'); copied.push(text); } } },
    require(name) {
      if (name === 'react') return { ...React, useState: () => [[confirmed, brief][stateIndex++], () => {}] };
      if (name === 'react/jsx-runtime') return jsxRuntime;
      if (name === '@tanstack/react-query') return {
        useQueryClient: () => ({ invalidateQueries: async () => {} }),
        useQuery: options => {
          queryOptions = options;
          return { data: error ? undefined : [{ id: 'gemini', name: 'Google Gemini', configured, active: false }], isPending: false, isError: error, error: new Error('API indisponível'), refetch() {} };
        },
        useMutation: options => {
          mutationOptions = options;
          return { isPending: false, isError: false, mutate: value => options.mutationFn(value) };
        },
      };
      if (name === '@/shared/api-client/endpoints') return { api: { listFactoryTools() {}, async selectFactoryTool(...args) { calls.push(args); return { success: true }; } } };
      throw Error(name);
    },
  };
  vm.runInNewContext(code, context);
  const tree = componentModule.exports.FactoryToolsPanel({ isDemoMode: demo, onShowToast: message => messages.push(message) });
  function find(id, node = tree) {
    if (!node || typeof node !== 'object') return;
    if (node.props?.id === id) return node;
    for (const child of React.Children.toArray(node.props?.children)) { const found = find(id, child); if (found) return found; }
  }
  return { tree, find, calls, copied, messages, queryOptions, mutationOptions };
}

test('provider activation requires free tier confirmation and a configured key', () => {
  assert.equal(panel().find('factory-tool-gemini').props.disabled, true);
  assert.equal(panel({ confirmed: true, configured: false }).find('factory-tool-gemini').props.disabled, true);
  const h = panel({ confirmed: true });
  assert.equal(h.find('factory-tool-gemini').props.disabled, false);
  h.find('factory-tool-gemini').props.onClick();
  assert.deepEqual(h.calls, [['gemini', true, true]]);
});

test('API failure does not fabricate a connection or activation control', () => {
  const h = panel({ error: true });
  assert.equal(h.find('factory-tool-gemini'), undefined);
  assert.match(renderToStaticMarkup(h.tree), /Não foi possível consultar/);
});

test('demo disables real configuration reads and blocks mutation even if invoked directly', async () => {
  const h = panel({ demo: true, confirmed: true });
  assert.equal(h.queryOptions.enabled, false);
  assert.equal(h.find('factory-tool-gemini').props.disabled, true);
  await assert.rejects(async () => h.mutationOptions.mutationFn({ id: 'gemini', active: true }), /Entre na sua conta/);
  assert.equal(h.calls.length, 0);
});

test('Figma brief is copied manually without starting an API generation', async () => {
  const h = panel({ brief: 'Agência de veículos' });
  await h.find('figma-copy-brief').props.onClick();
  assert.match(h.copied[0], /Agência de veículos/);
  assert.match(h.copied[0], /360 px/);
  assert.equal(h.calls.length, 0);
  assert.equal(h.find('figma-open-make').props.href, 'https://www.figma.com/make/');
  const failed = panel({ brief: 'Loja', clipboardFails: true });
  await failed.find('figma-copy-brief').props.onClick();
  assert.match(failed.messages[0], /Não foi possível copiar/);
});
