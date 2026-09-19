import {cpSync, existsSync, mkdirSync, readFileSync, readdirSync, rmSync, statSync, writeFileSync} from 'node:fs';
import {join, posix} from 'node:path';
import {fileURLToPath} from 'node:url';

const docsRoot = fileURLToPath(new URL('../', import.meta.url));
const repoRoot = fileURLToPath(new URL('../../', import.meta.url));
const input = join(docsRoot, 'build/docs-input');
const output = join(docsRoot, 'build/site');
const ref = process.env.DOCS_SOURCE_REF || 'new';
const github = 'https://github.com/m8-team/platform';

// Build from an allowlist, never from the whole repository or installed skills.
rmSync(input, {recursive: true, force: true});
rmSync(output, {recursive: true, force: true});
mkdirSync(join(input, 'docs'), {recursive: true});

for (const entry of ['README.md', 'principles.md', 'glossary.md', 'architecture', 'adr', 'development', 'specs']) {
  cpSync(join(docsRoot, entry), join(input, 'docs', entry), {
    recursive: true,
    filter: (path) => !path.split('/').some((part) => part.startsWith('.') && part !== '.'),
  });
}

cpSync(join(docsRoot, 'toc.yaml'), join(input, 'toc.yaml'));
cpSync(join(repoRoot, 'README.md'), join(input, 'README.md'));
cpSync(join(repoRoot, 'AGENTS.md'), join(input, 'AGENTS.md'));

const published = (path) => path === 'README.md' ? 'index.md' : path;
let sourceLinks = 0;
let pages = 0;

function resolveLink(source, href) {
  if (/^(?:[a-z][a-z\d+.-]*:|\/\/|#)/i.test(href)) return href;
  const [, pathname, suffix = ''] = href.match(/^([^?#]*)(.*)$/);
  if (!pathname) return href;
  const target = posix.normalize(posix.join(posix.dirname(source), decodeURI(pathname)));
  if (pathname.startsWith('/') || target.startsWith('../')) {
    throw new Error(`${source}: use a repository-relative link: ${href}`);
  }
  const staged = join(input, target);
  if (existsSync(staged) && posix.basename(target) !== 'toc.yaml') {
    let document = target;
    if (statSync(staged).isDirectory()) {
      document = ['index.md', 'README.md'].map((name) => posix.join(target, name))
        .find((path) => existsSync(join(input, path)));
    }
    if (document) {
      return (posix.relative(posix.dirname(published(source)), published(document)) || '.') + suffix;
    }
  }
  const original = join(repoRoot, target);
  if (!existsSync(original)) throw new Error(`${source}: missing link target: ${href}`);
  sourceLinks++;
  const kind = statSync(original).isDirectory() ? 'tree' : 'blob';
  return `${github}/${kind}/${encodeURIComponent(ref)}/${target.split('/').map(encodeURIComponent).join('/')}${suffix}`;
}

function visit(directory) {
  for (const entry of readdirSync(directory, {withFileTypes: true})) {
    const file = join(directory, entry.name);
    if (entry.isDirectory()) { visit(file); continue; }
    if (!entry.name.endsWith('.md')) continue;
    const source = file.slice(input.length + 1);
    let fence;
    const content = readFileSync(file, 'utf8').split('\n').map((line) => {
      const marker = line.match(/^\s{0,3}(`{3,}|~{3,})/);
      if (marker) {
        if (!fence) fence = marker[1];
        else if (marker[1][0] === fence[0] && marker[1].length >= fence.length) fence = undefined;
        return line;
      }
      if (fence) return line;
      return line.split(/(`+[^`]*`+)/g).map((part) => {
        if (part.startsWith('`')) return part;
        return part.replace(/(!?\[[^\]\n]*\]\()([^\s)]+)([^)\n]*\))/g,
          (_, start, href, end) => start + resolveLink(source, href) + end)
          .replace(/^(\s{0,3}\[[^\]]+\]:\s*)(\S+)(.*)$/,
            (_, start, href, end) => start + resolveLink(source, href) + end);
      }).join('');
    }).join('\n');
    writeFileSync(file, content);
    pages++;
  }
}

visit(input);
cpSync(join(input, 'README.md'), join(input, 'index.md'));
rmSync(join(input, 'README.md'));
console.log(`Prepared ${pages} pages; ${sourceLinks} source links point to GitHub (${ref}).`);
