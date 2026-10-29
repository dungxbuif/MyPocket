# iOS Personal Team signing

The native app keeps `applinks:money.dungxbuif.com` in the production entitlements for Universal Links. Apple Personal Teams do not support the Associated Domains capability, so a locally installed development/release build signed with the personal Team ID must use the empty entitlement file at `mobile-ios/MyPocketIOS/Signing/MyPocketIOS.personal.entitlements`.

This fallback only removes the Universal Links entitlement from that local build. The app's custom URL-scheme routes and auth/route guards remain available. App Store or paid-team builds must continue using `mobile-ios/MyPocketIOS/MyPocketIOS.entitlements`.

Example device build override:

```sh
xcodebuild build \
  -project MyPocketIOS.xcodeproj \
  -scheme MyPocketIOS \
  -destination 'id=<device-udid>' \
  DEVELOPMENT_TEAM=AU442S3MN3 \
  CODE_SIGN_IDENTITY='Apple Development' \
  CODE_SIGN_ENTITLEMENTS=MyPocketIOS/Signing/MyPocketIOS.personal.entitlements
```
