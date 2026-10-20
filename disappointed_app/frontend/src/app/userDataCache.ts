const USER_DATA_CACHE_PREFIX = "mypocket:user-data-cache:v1:";

export function userDataCacheKey(userID: string, resource: string) {
  return `${USER_DATA_CACHE_PREFIX}${encodeURIComponent(userID)}:${resource}`;
}

export function readUserDataCache<T>(userID: string, resource: string): T | null {
  try {
    const value = localStorage.getItem(userDataCacheKey(userID, resource));
    return value ? JSON.parse(value) as T : null;
  } catch {
    return null;
  }
}

export function writeUserDataCache(userID: string, resource: string, value: unknown) {
  try { localStorage.setItem(userDataCacheKey(userID, resource), JSON.stringify(value)); }
  catch { /* Cache is optional; preserve a successful online response. */ }
}

export function clearUserDataCaches(userID: string) {
  const prefix = `${USER_DATA_CACHE_PREFIX}${encodeURIComponent(userID)}:`;
  Object.keys(localStorage)
    .filter((key) => key.startsWith(prefix))
    .forEach((key) => localStorage.removeItem(key));
}
