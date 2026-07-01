export type A2UIScreen = {
  id: string
  route: string
  title: string
  surface: {
    id: string
    component: string
    props?: Record<string, unknown>
    children?: A2UIComponent[]
  }
  dataModel: Record<string, unknown>
  actions: A2UIAction[]
}

export type A2UIComponent = {
  id: string
  component: string
  props?: Record<string, unknown>
  dataPath?: string
  children?: A2UIComponent[]
}

export type A2UIAction = {
  name: string
  type: 'event' | 'local'
  payload?: Record<string, unknown>
}
