import { expect,test } from "vitest";
import { effectiveAgentStatus, type AgentRunStatus, type AgentRunView } from "./agent";

for (const status of ["queued","submitting","processing","completed","failed","cancelled","expired"] satisfies AgentRunStatus[]) {
  test(`exposes ${status} image-tool state`,()=>{
    const run:AgentRunView={id:"r",kind:"analysis",status:"queued",request_text:"x",attempts:0,draft_ids:[],tool_runs:[{id:"t",status,attempts:1}]};
    expect(effectiveAgentStatus(run)).toBe(status==="completed"?"queued":status);
  });
}
