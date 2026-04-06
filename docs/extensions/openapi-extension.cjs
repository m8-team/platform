const path = require('node:path');

const {getHooks: getProgramHooks} = require('@diplodoc/cli/lib/program');
const {getHooks: getTocHooks} = require('@diplodoc/cli/lib/toc');
const {normalizePath} = require('@diplodoc/cli/lib/utils');
const {includer} = require('@diplodoc/openapi-extension/includer');

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

          service.relations.addNode(input, {type: 'generator', data: undefined});
          service.relations.addNode(rawtoc.path, {type: 'source', data: undefined});
          service.relations.addDependency(rawtoc.path, input);

          const {toc, files} = await includer(run, options, from);
          const root = path.join(run.input, path.dirname(options.path));
          const maxOpenapiIncludeSize = run.config?.content?.maxOpenapiIncludeSize || 0;

          for (const {path: filePath, content} of files) {
            if (
              maxOpenapiIncludeSize > 0 &&
              Buffer.byteLength(content, 'utf8') > maxOpenapiIncludeSize
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
                  `(${Buffer.byteLength(content, 'utf8')} > ${maxOpenapiIncludeSize} bytes). ` +
                  'Replacing with stub.',
              );
              await run.write(path.join(root, filePath), stub, true);
              continue;
            }

            await run.write(path.join(root, filePath), content, true);
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

module.exports = {Extension};
