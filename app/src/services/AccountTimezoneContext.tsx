import { createContext, useContext } from "react";
import { DEFAULT_ACCOUNT_TIMEZONE } from "./accountTime";

export { DEFAULT_ACCOUNT_TIMEZONE };
const AccountTimezoneContext = createContext(DEFAULT_ACCOUNT_TIMEZONE);

export const AccountTimezoneProvider = AccountTimezoneContext.Provider;
export function useAccountTimezone(): string {
  return useContext(AccountTimezoneContext);
}
