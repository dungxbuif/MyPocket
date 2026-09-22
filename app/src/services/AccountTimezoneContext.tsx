import { createContext, useContext } from "react";

export const DEFAULT_ACCOUNT_TIMEZONE = "Asia/Ho_Chi_Minh";
const AccountTimezoneContext = createContext(DEFAULT_ACCOUNT_TIMEZONE);

export const AccountTimezoneProvider = AccountTimezoneContext.Provider;
export function useAccountTimezone(): string {
  return useContext(AccountTimezoneContext);
}
