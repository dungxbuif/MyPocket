import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe,expect,it } from "vitest";

const mountedSurfaces=["src/app/App.tsx","src/screens/OverviewScreen.tsx","src/screens/TransactionsScreen.tsx","src/screens/BudgetsScreen.tsx","src/screens/AccountScreen.tsx","src/screens/AgentScreen.tsx"];

describe("mounted UI contract",()=>{
  it("uses shared interactive primitives instead of screen-local native controls",()=>{
    for(const file of mountedSurfaces){const source=readFileSync(join(process.cwd(),file),"utf8");expect(source,`${file} contains a raw interactive control`).not.toMatch(/<(button|input|select|textarea)\b/)}
  });
  it("does not reintroduce removed chromatic product accents",()=>{
    const banned=/#(?:2dbd4f|39c85f|14b8a6|2f80ed|32a9df|7f5af0|ff5a66|ff1717|ff8800)\b/i;
    for(const file of mountedSurfaces){expect(readFileSync(join(process.cwd(),file),"utf8"),file).not.toMatch(banned)}
  });
});
