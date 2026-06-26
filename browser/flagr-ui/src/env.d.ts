/// <reference types="vite/client" />

import type { ElementMessageApi } from '@/ui/runApi'

interface ImportMetaEnv {
  readonly VITE_API_URL?: string
  readonly VITE_FLAGR_UI_POSSIBLE_ENTITY_TYPES?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

declare module 'vue' {
  interface ComponentCustomProperties {
    $message: ElementMessageApi
    $confirm: (
      message: string,
      title: string,
      options?: Record<string, unknown>,
    ) => Promise<void>
  }
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<object, object, unknown>
  export default component
}

declare module '*.json' {
  const value: unknown
  export default value
}

declare module 'markdown-it' {
  class MarkdownIt {
    constructor(options?: unknown)
    render(src: string): string
  }
  export default MarkdownIt
}