import { readFileSync, readdirSync, existsSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
const root = fileURLToPath(new URL('../../docs/design/', import.meta.url));
const walk = dir => readdirSync(dir, {withFileTypes:true}).flatMap(entry => entry.isDirectory() ? walk(path.join(dir,entry.name)) : [path.join(dir,entry.name)]);
const errors=[]; let links=0;
for(const file of walk(root)) {
  if(!file.endsWith('.md')) {
    const spec=path.join(path.dirname(file),'README.md');
    const declared=existsSync(spec) && /^artifact_source: visual_reference$/m.test(readFileSync(spec,'utf8'));
    if(!declared || !['code.html','screen.png'].includes(path.basename(file))) errors.push(`Non-spec artifact: ${file}`);
    continue;
  }
  for(const match of readFileSync(file,'utf8').matchAll(/\[[^\]]*\]\(([^)]+)\)/g)) {
    const target=match[1].split('#')[0];
    if(!target || /^[a-z]+:/i.test(target)) continue;
    links++; if(!existsSync(path.resolve(path.dirname(file),target))) errors.push(`${file}: missing ${target}`);
  }
}
if(errors.length) {console.error(errors.join('\n'));process.exitCode=1;} else console.log(`Design specs passed: ${links} valid local links; only declared reference assets allowed.`);
