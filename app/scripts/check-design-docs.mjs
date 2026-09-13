import { readFileSync, readdirSync, existsSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
const root = fileURLToPath(new URL('../../docs/design/', import.meta.url));
const walk = dir => readdirSync(dir, {withFileTypes:true}).flatMap(entry => entry.isDirectory() ? walk(path.join(dir,entry.name)) : [path.join(dir,entry.name)]);
const errors=[]; let links=0;
for(const file of walk(root)) {
  if(!file.endsWith('.md')) { errors.push(`Non-spec artifact: ${file}`); continue; }
  for(const match of readFileSync(file,'utf8').matchAll(/\[[^\]]*\]\(([^)]+)\)/g)) {
    const target=match[1].split('#')[0];
    if(!target || /^[a-z]+:/i.test(target)) continue;
    links++; if(!existsSync(path.resolve(path.dirname(file),target))) errors.push(`${file}: missing ${target}`);
  }
}
if(errors.length) {console.error(errors.join('\n'));process.exitCode=1;} else console.log(`Design specs passed: ${walk(root).length} Markdown files, ${links} valid local links, no PNG/HTML exports.`);
