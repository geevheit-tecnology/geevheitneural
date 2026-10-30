const postcss = require('postcss');
const tailwind = require('@tailwindcss/postcss');
const fs = require('fs');
const path = require('path');

const cssPath = path.join(__dirname, '../src/app/globals.css');
const outPath = path.join(__dirname, '../src/app/compiled.css');

if (fs.existsSync(cssPath)) {
  const css = fs.readFileSync(cssPath, 'utf8');
  postcss([tailwind()]).process(css, { from: cssPath }).then(result => {
    fs.writeFileSync(outPath, result.css);
    console.log('✅ Tailwind CSS compilado com sucesso:', result.css.length, 'bytes');
  }).catch(err => {
    console.error('❌ Erro na compilação do Tailwind CSS:', err);
    process.exit(1);
  });
} else {
  console.error('❌ globals.css não encontrado em:', cssPath);
  process.exit(1);
}
