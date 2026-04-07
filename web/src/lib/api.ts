const BASE_URL = '/api/v1'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `Request failed: ${res.status}`)
  }
  return res.json()
}

// Connection types
export interface Connection {
  id: string
  name: string
  url: string
  authMethod: string
  username?: string
  password?: string
  token?: string
  nkey?: string
  credsFile?: string
  monitorUrl: string
  createdAt: string
  updatedAt: string
}

export interface ConnectionStatus {
  connected: boolean
  connectionId?: string
  serverName?: string
  serverId?: string
  clusterName?: string
  version?: string
  rtt?: string
  error?: string
}

// Connections API
export const connectionsApi = {
  list: () => request<Connection[]>('/connections'),
  get: (id: string) => request<Connection>(`/connections/${id}`),
  create: (data: Partial<Connection>) =>
    request<Connection>('/connections', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: Partial<Connection>) =>
    request<Connection>(`/connections/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) =>
    request<{ status: string }>(`/connections/${id}`, { method: 'DELETE' }),
  connect: (id: string) =>
    request<ConnectionStatus>(`/connections/${id}/connect`, { method: 'POST' }),
  disconnect: (id: string) =>
    request<ConnectionStatus>(`/connections/${id}/disconnect`, { method: 'POST' }),
  status: () => request<ConnectionStatus>('/connection/status'),
}

// Server monitoring API
export const serverApi = {
  varz: () => request<any>('/server/varz'),
  jsz: () => request<any>('/server/jsz'),
  connz: () => request<any>('/server/connz'),
  healthz: () => request<any>('/server/healthz'),
  accountz: () => request<any>('/server/accountz'),
  accountInfo: () => request<any>('/server/account-info'),
}

// KV Store types
export interface KvBucket {
  name: string
  description: string
  values: number
  bytes: number
  history: number
  ttl: number
  storage: string
  replicas: number
  isCompressed: boolean
}

export interface KvEntry {
  key: string
  value: string
  valueText?: string
  revision: number
  delta: number
  created: string
  operation: string
}

// Stream types
export interface StreamInfo {
  config: StreamConfig
  state: StreamState
  created: string
}
export interface StreamConfig {
  name: string
  description: string
  subjects: string[]
  retention: string
  storage: string
  replicas: number
  maxMsgs: number
  maxBytes: number
  maxAge: number  // nanoseconds
  maxMsgSize: number
  maxConsumers: number
  maxMsgsPerSubject: number
  discard: string
  discardNewPerSubject: boolean
  duplicates: number
  noAck: boolean
  compression: string
  allowRollup: boolean
  denyDelete: boolean
  denyPurge: boolean
  allowDirect: boolean
  metadata: Record<string, string>
  sealed: boolean
  sources?: Array<{
    name: string
    filter_subject?: string
    opt_start_seq?: number
    opt_start_time?: string
    external?: { api: string; deliver: string }
    subject_transforms?: Array<{ src: string; dest: string }>
  }>
  mirror?: {
    name: string
    filter_subject?: string
    opt_start_seq?: number
    opt_start_time?: string
    external?: { api: string; deliver: string }
    subject_transforms?: Array<{ src: string; dest: string }>
  }
  mirror_direct?: boolean
}
export interface StreamState {
  messages: number
  bytes: number
  first_seq: number
  last_seq: number
  first_ts: string
  last_ts: string
  consumer_count: number
  num_subjects: number
  num_deleted: number
}
export interface StreamMessage {
  sequence: number
  subject: string
  data: string
  dataText?: string
  headers?: Record<string, string[]>
  time: string
  size: number
}
export interface ConsumerInfo {
  name: string
  config: any
  delivered: { consumer_seq: number; stream_seq: number }
  ack_floor: { consumer_seq: number; stream_seq: number }
  num_ack_pending: number
  num_redelivered: number
  num_waiting: number
  num_pending: number
  paused: boolean
  pause_remaining?: number
  created: string
}

export const streamsApi = {
  list: () => request<StreamInfo[]>('/streams'),
  get: (name: string) => request<StreamInfo>(`/streams/${name}`),
  create: (data: any) => request<StreamInfo>('/streams', { method: 'POST', body: JSON.stringify(data) }),
  update: (name: string, data: any) => request<StreamInfo>(`/streams/${name}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (name: string) => request<{ status: string }>(`/streams/${name}`, { method: 'DELETE' }),
  purge: (name: string, opts?: { subject?: string; seq?: number; keep?: number }) => {
    const params = new URLSearchParams()
    if (opts?.subject) params.set('subject', opts.subject)
    if (opts?.seq) params.set('seq', String(opts.seq))
    if (opts?.keep) params.set('keep', String(opts.keep))
    const qs = params.toString()
    return request<{ status: string }>(`/streams/${name}/purge${qs ? '?' + qs : ''}`, { method: 'POST' })
  },
  listMessages: (stream: string, opts?: { startSeq?: number; limit?: number; subject?: string; startDate?: string }) => {
    const params = new URLSearchParams()
    if (opts?.startSeq) params.set('startSeq', String(opts.startSeq))
    if (opts?.limit) params.set('limit', String(opts.limit))
    if (opts?.subject) params.set('subject', opts.subject)
    if (opts?.startDate) params.set('startDate', opts.startDate)
    const qs = params.toString()
    return request<{ messages: StreamMessage[]; firstSeq: number; lastSeq: number; total: number }>(`/streams/${stream}/messages${qs ? '?' + qs : ''}`)
  },
  getMessage: (stream: string, seq: number) => request<StreamMessage>(`/streams/${stream}/messages/${seq}`),
  getLastBySubject: (stream: string, subject: string) =>
    request<StreamMessage>(`/streams/${stream}/messages/last?subject=${encodeURIComponent(subject)}`),
  deleteMessage: (stream: string, seq: number) => request<{ status: string }>(`/streams/${stream}/messages/${seq}`, { method: 'DELETE' }),
}

export const consumersApi = {
  list: (stream: string) => request<ConsumerInfo[]>(`/streams/${stream}/consumers`),
  get: (stream: string, consumer: string) => request<ConsumerInfo>(`/streams/${stream}/consumers/${consumer}`),
  delete: (stream: string, consumer: string) => request<{ status: string }>(`/streams/${stream}/consumers/${consumer}`, { method: 'DELETE' }),
  pause: (stream: string, consumer: string, pauseUntil: string) => request<any>(`/streams/${stream}/consumers/${consumer}/pause`, { method: 'POST', body: JSON.stringify({ pauseUntil }) }),
  resume: (stream: string, consumer: string) => request<any>(`/streams/${stream}/consumers/${consumer}/resume`, { method: 'POST' }),
}

// Publish API
export const publishApi = {
  publish: (data: { subject: string; data?: string; headers?: Record<string, string[]> }) =>
    request<{ status: string }>('/publish', { method: 'POST', body: JSON.stringify(data) }),
  request: (data: { subject: string; data?: string; headers?: Record<string, string[]>; timeout?: string }) =>
    request<{ subject: string; data: string; dataText?: string; headers?: Record<string, string[]> }>('/request', { method: 'POST', body: JSON.stringify(data) }),
}

// Object Store types
export interface ObjectStoreInfo {
  bucket: string
  description: string
  sealed: boolean
  size: number
  storage: string
  replicas: number
  metadata: Record<string, string>
}

export interface ObjectInfo {
  name: string
  description: string
  size: number
  chunks: number
  digest: string
  modified: string
  deleted: boolean
  headers: Record<string, string[]>
}

// Object Store API
export const objectsApi = {
  listStores: () => request<ObjectStoreInfo[]>('/objects/stores'),
  createStore: (data: any) => request<ObjectStoreInfo>('/objects/stores', { method: 'POST', body: JSON.stringify(data) }),
  getStore: (name: string) => request<ObjectStoreInfo>(`/objects/stores/${name}`),
  deleteStore: (name: string) => request<{ status: string }>(`/objects/stores/${name}`, { method: 'DELETE' }),
  listObjects: (store: string) => request<ObjectInfo[]>(`/objects/stores/${store}/objects`),
  getObjectInfo: (store: string, name: string) => request<ObjectInfo>(`/objects/stores/${store}/objects/${encodeURIComponent(name)}`),
  deleteObject: (store: string, name: string) => request<{ status: string }>(`/objects/stores/${store}/objects/${encodeURIComponent(name)}`, { method: 'DELETE' }),
}

// KV Store API
export const kvApi = {
  listBuckets: () => request<KvBucket[]>('/kv/buckets'),
  getBucket: (name: string) => request<KvBucket>(`/kv/buckets/${name}`),
  createBucket: (data: any) => request<KvBucket>('/kv/buckets', { method: 'POST', body: JSON.stringify(data) }),
  deleteBucket: (name: string) => request<{ status: string }>(`/kv/buckets/${name}`, { method: 'DELETE' }),
  listKeys: (bucket: string, q?: string) => request<string[]>(`/kv/buckets/${bucket}/keys${q ? `?q=${encodeURIComponent(q)}` : ''}`),
  getKey: (bucket: string, key: string) => request<KvEntry>(`/kv/buckets/${bucket}/keys/${encodeURIComponent(key)}`),
  putKey: (bucket: string, key: string, value: string) => request<KvEntry>(`/kv/buckets/${bucket}/keys/${encodeURIComponent(key)}`, { method: 'PUT', body: JSON.stringify({ value }) }),
  deleteKey: (bucket: string, key: string) => request<{ status: string }>(`/kv/buckets/${bucket}/keys/${encodeURIComponent(key)}`, { method: 'DELETE' }),
  purgeKey: (bucket: string, key: string) => request<{ status: string }>(`/kv/buckets/${bucket}/keys/${encodeURIComponent(key)}/purge`, { method: 'POST' }),
  getKeyHistory: (bucket: string, key: string) => request<KvEntry[]>(`/kv/buckets/${bucket}/keys/${encodeURIComponent(key)}/history`),
}
