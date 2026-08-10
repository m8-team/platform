'use client';

import {
    createContext,
    useContext,
    type Dispatch,
    type SetStateAction,
} from 'react';

export type AppTheme = 'light' | 'dark';

interface AppThemeContextValue {
    theme: AppTheme;
    setTheme: Dispatch<SetStateAction<AppTheme>>;
}

export const AppThemeContext = createContext<AppThemeContextValue | null>(null);

export function useAppTheme() {
    const context = useContext(AppThemeContext);

    if (!context) {
        throw new Error('useAppTheme must be used within AppThemeContext.Provider');
    }

    return context;
}
