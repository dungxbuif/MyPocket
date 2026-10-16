# Mobile MCP physical-device probe — 2026-10-16

Status: verified connection and UI-read capability. Owner requested testing
[mobile-next/mobile-mcp](https://github.com/mobile-next/mobile-mcp) against the
connected physical iPhone during the Savings/Budget implementation.

## Intended operation

Run a pinned temporary Mobile MCP server locally with telemetry disabled.
Probe device discovery first, then screen/element inspection for the explicitly
selected MyPocket device. Do not inspect unrelated applications or change
financial data. No global MCP configuration change is required for this probe.

## Initial evidence

`xcrun devicectl list devices` and `device info details` detect the connected
iPhone 15 Plus, paired over USB with Developer Mode enabled.
`device info apps` finds MyPocket `com.mypocket.ios`, version 1.1.0 build 2.
No WebDriverAgent application is installed. npm reports Mobile MCP 1.0.11.
The maintained real-device setup guide requires a signed WebDriverAgent runner,
go-ios tunnel and forwarding; actual package behavior will be probed before
making a capability claim.

## Verification plan

Initialize the MCP transport, inspect advertised tool schemas, call device
discovery, and attempt a read-only screen/element probe using the returned
device identifier. Record exact success or prerequisite error. Installation
and signing of an automation runner is a separate platform operation if needed.

Human/public API docs need no change: this is developer tooling only. This
Markdown note is also the agent-readable operational reference.

## Verified outcome

Mobile MCP 1.0.11 initialized over local HTTP, discovered the physical phone
online and initially failed UI reads because the agent was absent. A dedicated
Personal Team profile was then created and Device Kit 0.0.34 installed under
`AU442S3MN3.com.mobilenext.devicekit-iosUITests.xctrunner`. MCP subsequently
read 58 home-screen elements; the survey agent launched and read Money Lover.
Independent devicectl inspection confirms MyPocket is still 1.1.0 build 2.

The helper profile expires 2026-10-23; this is free development signing, not a
permanent distribution install. [Installation command/evidence](IOS-AUTOMATION-INSTALL-20261016.md).
The temporary MCP listener binds 127.0.0.1:19347, with telemetry disabled; no
global Codex configuration was changed. Physical UI surveying is separately
tracked in the [reference ledger](../design/reference-survey/2026-10-16-money-lover-iphone.md).
