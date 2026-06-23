import {useEffect, useMemo, useState} from 'react'
import {configure, SegmentedRadioGroup, Text, ThemeProvider} from '@gravity-ui/uikit'
import {Table, useTable} from '@gravity-ui/table'
import type {ColumnDef} from '@gravity-ui/table/tanstack'

type AppLanguage = 'ru' | 'en'
type AppTheme = 'light' | 'dark'

interface Person {
  id: string
  name: string
  age: number
}

const initialLanguage: AppLanguage = 'ru'
const initialTheme: AppTheme = 'light'

const messages = {
  ru: {
    title: 'Пользователи',
    subtitle: 'Пример таблицы Gravity UI с переключателями языка и темы.',
    language: 'Язык',
    theme: 'Тема',
    light: 'Светлая',
    dark: 'Темная',
    name: 'Имя',
    age: 'Возраст',
  },
  en: {
    title: 'Users',
    subtitle: 'Gravity UI table example with language and theme switchers.',
    language: 'Language',
    theme: 'Theme',
    light: 'Light',
    dark: 'Dark',
    name: 'Name',
    age: 'Age',
  },
} satisfies Record<AppLanguage, Record<string, string>>

const data: Person[] = [
  {id: 'john', name: 'John', age: 23},
  {id: 'michael', name: 'Michael', age: 27},
]

configure({
  lang: initialLanguage,
  fallbackLang: 'en',
})

function App() {
  const [language, setLanguage] = useState<AppLanguage>(initialLanguage)
  const [theme, setTheme] = useState<AppTheme>(initialTheme)
  const t = messages[language]

  useEffect(() => {
    configure({
      lang: language,
      fallbackLang: 'en',
    })
    document.documentElement.lang = language
  }, [language])

  const columns = useMemo<ColumnDef<Person>[]>(
    () => [
      {accessorKey: 'name', header: t.name, size: 160},
      {accessorKey: 'age', header: t.age, size: 120},
    ],
    [t.age, t.name],
  )

  const table = useTable({
    columns,
    data,
  })

  return (
    <ThemeProvider theme={theme} lang={language} fallbackLang="en">
      <main
        style={{
          minHeight: '100vh',
          background: 'var(--g-color-base-background)',
          color: 'var(--g-color-text-primary)',
          padding: 32,
        }}
      >
        <section
          style={{
            display: 'flex',
            maxWidth: 920,
            flexDirection: 'column',
            gap: 24,
          }}
        >
          <header
            style={{
              display: 'flex',
              alignItems: 'flex-start',
              justifyContent: 'space-between',
              gap: 24,
            }}
          >
            <div>
              <Text as="h1" variant="header-2">
                {t.title}
              </Text>
              <Text as="p" variant="body-2" color="secondary">
                {t.subtitle}
              </Text>
            </div>

            <div
              style={{
                display: 'flex',
                flexWrap: 'wrap',
                justifyContent: 'flex-end',
                gap: 16,
              }}
            >
              <div style={{display: 'grid', gap: 8}}>
                <Text variant="caption-2" color="secondary">
                  {t.language}
                </Text>
                <SegmentedRadioGroup<AppLanguage>
                  size="m"
                  value={language}
                  onUpdate={setLanguage}
                  options={[
                    {value: 'ru', content: 'RU'},
                    {value: 'en', content: 'EN'},
                  ]}
                />
              </div>

              <div style={{display: 'grid', gap: 8}}>
                <Text variant="caption-2" color="secondary">
                  {t.theme}
                </Text>
                <SegmentedRadioGroup<AppTheme>
                  size="m"
                  value={theme}
                  onUpdate={setTheme}
                  options={[
                    {value: 'light', content: t.light},
                    {value: 'dark', content: t.dark},
                  ]}
                />
              </div>
            </div>
          </header>

          <div
            style={{
              overflow: 'hidden',
              border: '1px solid var(--g-color-line-generic)',
              borderRadius: 8,
              background: 'var(--g-color-base-background)',
            }}
          >
            <Table table={table} />
          </div>
        </section>
      </main>
    </ThemeProvider>
  )
}

export default App
