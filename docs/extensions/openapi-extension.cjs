const path = require('node:path');
const fs = require('node:fs');

const {getHooks: getProgramHooks} = require('@diplodoc/cli/lib/program');
const {getHooks: getTocHooks} = require('@diplodoc/cli/lib/toc');
const {normalizePath} = require('@diplodoc/cli/lib/utils');
const {includer} = require('@diplodoc/openapi-extension/includer');
const yaml = require('js-yaml');

const EXTENSION = 'OpenapiIncluder';
const INCLUDER = 'openapi';

class Extension {
  apply(program) {
    getProgramHooks(program).BeforeAnyRun.tap(EXTENSION, (run) => {
      getTocHooks(run.toc).Includer.for(INCLUDER).tapPromise(
        EXTENSION,
        async (rawtoc, options, from) => {
          const input = normalizePath(options.input);
          const service = run.toc;
          const pageTitlePrefix = readPageTitlePrefix(path.join(run.input, input));

          service.relations.addNode(input, {type: 'generator', data: undefined});
          service.relations.addNode(rawtoc.path, {type: 'source', data: undefined});
          service.relations.addDependency(rawtoc.path, input);

          const {toc, files} = await includer(run, options, from);
          const methodPageTitles = new Map();
          collectMethodPageTitles(toc, pageTitlePrefix, methodPageTitles);
          const root = path.join(run.input, path.dirname(options.path));
          const maxOpenapiIncludeSize = run.config?.content?.maxOpenapiIncludeSize || 0;

          for (const {path: filePath, content} of files) {
            const normalizedFilePath = normalizePath(filePath);
            const updatedContent = updateMethodPageHeading(
              content,
              methodPageTitles.get(normalizedFilePath) ??
                methodPageTitles.get(normalizePath(path.basename(filePath))),
            );

            if (
              maxOpenapiIncludeSize > 0 &&
              Buffer.byteLength(updatedContent, 'utf8') > maxOpenapiIncludeSize
            ) {
              const stub = [
                '---',
                'noIndex: true',
                '---',
                '',
                '{% note warning %}',
                '',
                'This page exceeds the maximum allowed size and cannot be displayed.',
                '',
                '{% endnote %}',
              ].join('\n');

              run.logger.warn(
                `OpenAPI page ${filePath} exceeds max-openapi-include-size limit ` +
                  `(${Buffer.byteLength(updatedContent, 'utf8')} > ${maxOpenapiIncludeSize} bytes). ` +
                  'Replacing with stub.',
              );
              await run.write(path.join(root, filePath), stub, true);
              continue;
            }

            await run.write(path.join(root, filePath), updatedContent, true);
          }

          await service.walkEntries([toc], async (entry) => {
            const entryPath = normalizePath(path.join(path.dirname(options.path), entry.href));
            service.relations.addNode(entryPath, {type: 'entry', data: undefined});
            service.relations.addDependency(input, entryPath);

            return entry;
          });

          return toc;
        },
      );
    });
  }
}

function readPageTitlePrefix(openapiPath) {
  try {
    const spec = yaml.load(fs.readFileSync(openapiPath, 'utf8'));
    const title = spec?.info?.title;

    if (typeof title !== 'string' || title.trim().length === 0) {
      return 'API';
    }

    return normalizeApiTitle(title);
  } catch {
    return 'API';
  }
}

function normalizeApiTitle(title) {
  let normalized = title.trim().replace(/\s+Service(?=\s+API$|$)/, '');

  if (!/\sAPI$/.test(normalized)) {
    normalized = `${normalized} API`;
  }

  return normalized;
}

function collectMethodPageTitles(entry, pageTitlePrefix, methodPageTitles) {
  if (!entry || typeof entry !== 'object') {
    return;
  }

  const items = Array.isArray(entry.items) ? entry.items : [];

  if (items.length > 0) {
    for (const item of items) {
      collectMethodPageTitles(item, pageTitlePrefix, methodPageTitles);
    }

    return;
  }

  if (typeof entry.href !== 'string' || entry.href === 'index.md' || entry.href.endsWith('/index.md')) {
    return;
  }

  const pageTitle = `${pageTitlePrefix}: ${entry.name}`;
  methodPageTitles.set(normalizePath(entry.href), pageTitle);
  methodPageTitles.set(normalizePath(path.basename(entry.href)), pageTitle);
}

function updateMethodPageHeading(content, pageTitle) {
  if (!pageTitle) {
    return content;
  }

  return content.replace(/^#\s+.+$/m, `# ${pageTitle}`);
}

module.exports = {Extension};
