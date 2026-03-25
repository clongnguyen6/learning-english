import {
  createContext,
  useState,
  type ReactNode,
} from "react";

type ShellPreferencesValue = {
  compactRail: boolean;
  toggleCompactRail: () => void;
};

type ShellPreferencesProviderProps = {
  children: ReactNode;
};

export const ShellPreferencesContext =
  createContext<ShellPreferencesValue | null>(null);

export function ShellPreferencesProvider({
  children,
}: ShellPreferencesProviderProps) {
  const [compactRail, setCompactRail] = useState(false);

  return (
    <ShellPreferencesContext.Provider
      value={{
        compactRail,
        toggleCompactRail: () => {
          setCompactRail((currentValue) => !currentValue);
        },
      }}
    >
      {children}
    </ShellPreferencesContext.Provider>
  );
}
