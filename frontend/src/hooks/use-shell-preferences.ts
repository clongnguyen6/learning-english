import { useContext } from "react";
import { ShellPreferencesContext } from "../store/shell-preferences";

export function useShellPreferences() {
  const value = useContext(ShellPreferencesContext);

  if (!value) {
    throw new Error(
      "useShellPreferences must be used inside ShellPreferencesProvider.",
    );
  }

  return value;
}
