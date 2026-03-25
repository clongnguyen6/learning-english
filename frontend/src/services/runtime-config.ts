const DEFAULT_API_BASE_URL = "http://localhost:8080/api/v1";
const DEFAULT_APP_NAME = "learning-english";

type RuntimeConfig = {
  appEnv: string;
  appName: string;
  apiBaseUrl: string;
  appOrigin: string;
};

function normalizeValue(value: string | undefined, fallback: string) {
  const trimmed = value?.trim();

  return trimmed && trimmed.length > 0 ? trimmed : fallback;
}

export function getRuntimeConfig(): RuntimeConfig {
  return {
    appEnv: normalizeValue(import.meta.env.VITE_APP_ENV, "development"),
    appName: normalizeValue(import.meta.env.VITE_APP_NAME, DEFAULT_APP_NAME),
    apiBaseUrl: normalizeValue(
      import.meta.env.VITE_API_BASE_URL,
      DEFAULT_API_BASE_URL,
    ),
    appOrigin: normalizeValue(
      import.meta.env.VITE_APP_ORIGIN,
      window.location.origin,
    ),
  };
}
