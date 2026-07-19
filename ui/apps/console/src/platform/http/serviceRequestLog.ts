import type {ServiceRequestValues} from './serviceRequestCapture'

const MAXIMUM_RECORDS = 100

export interface ServiceRequestRecord {
  readonly id: string
  readonly service: string
  readonly method: string
  readonly url: string
  readonly parameters: ServiceRequestValues
  readonly requestHeaders: ServiceRequestValues
  readonly requestBody?: unknown
  readonly responseHeaders?: ServiceRequestValues
  readonly responseBody?: unknown
  readonly responseBodyPending?: boolean
  readonly startedAt: string
  readonly durationMs?: number
  readonly status?: number
  readonly error?: string
  readonly pending: boolean
}

export type ServiceRequestRecordUpdate = Partial<Omit<ServiceRequestRecord, 'id'>>
type Listener = () => void

class ServiceRequestStore {
  private records: readonly ServiceRequestRecord[] = []
  private readonly listeners = new Set<Listener>()

  subscribe = (listener: Listener) => {
    this.listeners.add(listener)
    return () => this.listeners.delete(listener)
  }

  getSnapshot = () => this.records

  clear = () => {
    if (this.records.length === 0) return
    this.records = []
    this.emitChange()
  }

  prepend(record: ServiceRequestRecord) {
    this.records = [record, ...this.records].slice(0, MAXIMUM_RECORDS)
    this.emitChange()
  }

  update(id: string, update: ServiceRequestRecordUpdate) {
    const index = this.records.findIndex((record) => record.id === id)
    if (index === -1) return

    this.records = this.records.map((record, recordIndex) =>
      recordIndex === index ? {...record, ...update} : record,
    )
    this.emitChange()
  }

  private emitChange() {
    for (const listener of this.listeners) {
      try {
        listener()
      } catch {
        // API diagnostics must never affect the instrumented request.
        continue
      }
    }
  }
}

const requestStore = new ServiceRequestStore()

export const serviceRequestLog = {
  subscribe: requestStore.subscribe,
  getSnapshot: requestStore.getSnapshot,
  clear: requestStore.clear,
}

export function prependServiceRequest(record: ServiceRequestRecord) {
  requestStore.prepend(record)
}

export function updateServiceRequest(id: string, update: ServiceRequestRecordUpdate) {
  requestStore.update(id, update)
}
