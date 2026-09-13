import ts from 'typescript';
import { readdirSync, readFileSync, existsSync } from 'node:fs';
import { dirname, join, relative, resolve, sep } from 'node:path';
import { root } from './env.mjs';

const sourceRoot = join(root, 'frontend/src');
const layers = ['shared', 'entities', 'features', 'widgets', 'pages', 'app'];
const errors = [];
const normalize = value => value.split(sep).join('/');
function walk(folder) {
  return readdirSync(folder, { withFileTypes: true }).flatMap(entry => entry.isDirectory() ? walk(join(folder, entry.name)) : [join(folder, entry.name)]);
}
function checkImport(file, specifier) {
  if (!specifier.startsWith('@/') && !specifier.startsWith('.')) return;
  const target = specifier.startsWith('@/') ? resolve(sourceRoot, specifier.slice(2)) : resolve(dirname(file), specifier);
  const from = normalize(relative(sourceRoot, file)).split('/');
  const to = normalize(relative(sourceRoot, target)).split('/');
  const fromLayer = layers.indexOf(from[0]);
  const toLayer = layers.indexOf(to[0]);
  const report = message => errors.push(`${normalize(relative(root, file))}: ${message}: ${specifier}`);
  if (toLayer === -1 || fromLayer === -1) return report('Unknown FSD layer');
  if (toLayer > fromLayer) return report('Import from an upper layer');
  if (toLayer === fromLayer) {
    if (['app', 'shared'].includes(from[0]) || from[1] === to[1]) return;
    return report('Import between sibling slices');
  }
  // Styles and static assets may be imported directly; TypeScript APIs are encapsulated.
  if (/\.(css|svg|woff2)$/.test(specifier)) return;
  if (to[0] !== 'shared' && to.length !== 2 && !(to.length === 3 && to[2] === 'index')) return report('Use the slice public API');
  if (!existsSync(join(target, 'index.ts')) && !existsSync(join(target, 'index.tsx')) && !target.endsWith('index')) return report('Public index.ts is missing');
}
for (const file of walk(sourceRoot).filter(file => /\.tsx?$/.test(file) && !file.includes(`${sep}generated${sep}`))) {
  const source = ts.createSourceFile(file, readFileSync(file, 'utf8'), ts.ScriptTarget.Latest, true);
  function visit(node) {
    if ((ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) && node.moduleSpecifier && ts.isStringLiteral(node.moduleSpecifier)) checkImport(file, node.moduleSpecifier.text);
    if (ts.isCallExpression(node) && node.expression.kind === ts.SyntaxKind.ImportKeyword && node.arguments[0] && ts.isStringLiteral(node.arguments[0])) checkImport(file, node.arguments[0].text);
    ts.forEachChild(node, visit);
  }
  visit(source);
}
if (errors.length) { console.error(errors.join('\n')); process.exit(1); }
console.log('FSD: layer boundaries and public APIs passed.');
