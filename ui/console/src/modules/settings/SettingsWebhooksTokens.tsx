import {useMemo,useState} from 'react'
import {ArrowRotateRight,Check,Magnifier,Shield,TriangleExclamation} from '@gravity-ui/icons'
import {Button,Icon,Text} from '@gravity-ui/uikit'

import {
  allowedWebhookEvents,requestDemoToken,retryDelivery,revokeDemoToken,
  saveWebhook,tokenScopes,toggleWebhook,useSettingsDemo,
} from './demo'
import type {ApiToken,TokenScope,Webhook,WebhookEvent,WebhookDelivery} from './demo'
import {Detail,Empty,Input,Metric,Modal,Notice,Panel,Pill,Select,SettingsPage} from './SettingsShell'
import type {T} from './SettingsShell'

function WebhookEditor({current,close,t}:{current?:Webhook;close:()=>void;t:T}){
  const [name,setName]=useState(current?.name??'')
  const [url,setUrl]=useState(current?.url??'https://hooks.example.com/events')
  const [events,setEvents]=useState<WebhookEvent[]>(current?.events??[])
  const [secretRef,setSecretRef]=useState(current?.secretRef??'')
  const [error,setError]=useState('')
  const toggle=(event:WebhookEvent)=>{
    setEvents(previous=>previous.includes(event)?previous.filter(e=>e!==event):[...previous,event])
  }
  return <Modal title={current?t('Изменить вебхук','Edit webhook'):t('Новый вебхук','Create webhook')} close={close}>
    <form className="st-form" onSubmit={e=>{
      e.preventDefault()
      const issue=saveWebhook({name,url,events,secretRef},current?.id)
      if(issue)setError(issue);else close()
    }}>
      <Input label={t('Имя вебхука','Webhook name')} value={name} onChange={setName} placeholder="Production incidents"/>
      <Input label="Delivery URL (HTTPS)" value={url} onChange={setUrl} placeholder="https://webhooks.example.com/m8"/>
      <Input label={t('Путь к ключу подписи в Vault','Signing key reference (Vault path)')} value={secretRef} onChange={setSecretRef} placeholder="webhooks/team-notifications"/>
      <div className="st-field">
        <span>{t('Типы событий','Event topics')}</span>
        <div className="st-checkbox-grid">{allowedWebhookEvents.map(event=><label key={event}><input type="checkbox" checked={events.includes(event)} onChange={()=>toggle(event)}/><span className="st-mono">{event}</span></label>)}</div>
      </div>
      <Notice tone="warning">{t('Новые вебхуки создаются отключёнными. Для включения требуется ссылка на ключ подписи. Никаких сетевых запросов и генерации секретов в демо нет.','New webhooks start disabled. Enabling requires a signing secret reference. Demo mode never makes network calls or issues secrets.')}</Notice>
      {error?<Notice tone="danger">{error}</Notice>:null}
      <div className="st-actions st-actions_end"><Button type="button" view="outlined" onClick={close}>{t('Отмена','Cancel')}</Button><Button type="submit" view="action">{t('Сохранить','Save demo')}</Button></div>
    </form>
  </Modal>
}
function WebhookDeliveryDetails({delivery,close,t}:{delivery:WebhookDelivery;close:()=>void;t:T}){
  const demo=useSettingsDemo()
  const [message,setMessage]=useState('')
  const webhook=demo.webhooks.find(w=>w.id===delivery.webhookId)
  const canRetry=delivery.status==='Failed'&&webhook?.status==='Enabled'
  const tryAgain=()=>{
    const id=retryDelivery(delivery.id)
    setMessage(id?t('Создана новая имитация доставки; HTTP-запрос не отправлен','A new simulated delivery was created; no HTTP request sent'):t('Повтор невозможен','Retry not available'))
  }
  return <Modal title={delivery.id} close={close}>
    <div className="st-details">
      <Detail label="Webhook" value={webhook?.name??delivery.webhookId}/>
      <Detail label="Event" value={<span className="st-mono">{delivery.event}</span>}/>
      <Detail label="Status" value={<Pill tone={delivery.status==='Failed'?'danger':delivery.status==='Delivered'?'positive':'warning'}>{delivery.status}</Pill>}/>
      <Detail label="Attempts" value={delivery.attempt}/>
      <Detail label="Response" value={delivery.responseCode??'No network request'}/>
      <Detail label={t('Время','Timestamp')} value={delivery.timestamp}/>
      <Detail label="Operation ID" value={delivery.operationId??'—'}/>
    </div>
    <Notice tone="warning">{t('Новый retry никогда не пересылает исходные данные. Демо сохраняет только новую запись о симуляции.','A retry never transmits the payload. Demo creates a new simulation record only.')}</Notice>
    {message?<Notice tone="info">{message}</Notice>:null}
    <div className="st-actions st-actions_end"><Button view="outlined" onClick={close}>{t('Закрыть','Close')}</Button>{canRetry?<Button view="action" onClick={tryAgain}><Icon data={ArrowRotateRight} size={15}/>{t('Повторить (демо)','Retry (demo)')}</Button>:null}</div>
  </Modal>
}
export function SettingsWebhooksPage(){
  const demo=useSettingsDemo()
  const [query,setQuery]=useState('')
  const [status,setStatus]=useState('All')
  const [creating,setCreating]=useState(false)
  const [editing,setEditing]=useState<string|null>(null)
  const [selected,setSelected]=useState<string|null>(null)
  const [selectedDelivery,setSelectedDelivery]=useState<string|null>(null)
  const [notice,setNotice]=useState('')
  const rows=useMemo(()=>demo.webhooks.filter(h=>(status==='All'||h.status===status)&&[h.name,h.url,h.id,...h.events].some(x=>x.toLowerCase().includes(query.toLowerCase()))),[demo.webhooks,status,query])
  const webhook=demo.webhooks.find(h=>h.id===selected)
  const currentEditor=demo.webhooks.find(h=>h.id===editing)
  const delivery=demo.deliveries.find(d=>d.id===selectedDelivery)
  return <SettingsPage tab="webhooks" title="Webhooks" subtitle="Event subscriptions, delivery attempts and signed outbound integrations">
    {t=><>
      <div className="st-section-heading"><div><Text as="h2" variant="header-2">{t('Исходящие вебхуки','Outbound webhooks')}</Text><Text color="secondary" variant="body-2">{t('Подписки на события M8 и история попыток доставки','M8 event subscriptions and delivery attempt history')}</Text></div>
        <Button view="action" onClick={()=>setCreating(true)}>+ {t('Создать вебхук','New webhook')}</Button>
      </div>
      <div className="st-kpis">
        <Metric label={t('Вебхуков','Webhooks')} value={demo.webhooks.length}/>
        <Metric label="Enabled" value={demo.webhooks.filter(h=>h.status==='Enabled').length}/>
        <Metric label={t('Доставлено','Delivered')} value={demo.deliveries.filter(d=>d.status==='Delivered').length} tone="positive"/>
        <Metric label={t('Ошибки доставки','Failed attempts')} value={demo.deliveries.filter(d=>d.status==='Failed').length} tone="warning"/>
      </div>
      <Panel title={t('Подписки','Subscriptions')} subtitle={t('Все действия локальны: отправка в Интернет отключена','Local demo only: outbound delivery is disabled')}>
        <div className="st-toolbar"><div className="st-search"><Icon data={Magnifier} size={16}/><input aria-label="Search webhooks" value={query} onChange={e=>setQuery(e.target.value)} placeholder={t('Имя, адрес или событие','Name, endpoint or event')}/></div><Select label="Status" value={status} onChange={setStatus} options={['All','Enabled','Disabled'].map(v=>({value:v,label:v}))}/></div>
        <div className="st-table-scroll"><table className="st-table"><thead><tr><th>Webhook</th><th>Endpoint</th><th>{t('События','Events')}</th><th>{t('Подпись','Signing')}</th><th>Status</th><th/></tr></thead><tbody>{rows.map(h=><tr key={h.id}>
          <td><strong>{h.name}</strong><span className="st-muted st-mono">{h.id}</span></td>
          <td className="st-mono">{h.url}</td><td>{h.events.length} {t('типов','topics')}</td>
          <td><Pill tone={h.signing==='External Secret'?'positive':'warning'}>{h.signing}</Pill></td>
          <td><Pill tone={h.status==='Enabled'?'positive':'neutral'}>{h.status}</Pill></td>
          <td><Button size="s" view="outlined" onClick={()=>setSelected(h.id)}>{t('Детали','Details')}</Button></td>
        </tr>)}</tbody></table></div>
        {!rows.length?<Empty>{t('Нет подходящих вебхуков','No matching webhooks')}</Empty>:null}
      </Panel>
      {notice?<Notice tone="warning">{notice}</Notice>:null}
      <Panel title={t('Попытки доставки','Delivery attempts')} subtitle={t('Включая ошибки, историю повторов и корреляцию с операциями','Statuses, retries and related operation IDs')}>
        <div className="st-table-scroll"><table className="st-table"><thead><tr><th>{t('Время','Timestamp')}</th><th>Webhook</th><th>Event</th><th>Status</th><th>{t('Попытки','Attempts')}</th><th>Response</th><th/></tr></thead><tbody>{demo.deliveries.map(d=><tr key={d.id}>
          <td className="st-nowrap">{d.timestamp}</td>
          <td>{demo.webhooks.find(h=>h.id===d.webhookId)?.name??d.webhookId}</td><td className="st-mono">{d.event}</td>
          <td><Pill tone={d.status==='Delivered'?'positive':d.status==='Failed'?'danger':'warning'}>{d.status}</Pill></td><td>{d.attempt}</td><td>{d.responseCode??'—'}</td>
          <td><Button size="s" view="outlined" onClick={()=>setSelectedDelivery(d.id)}>{t('Детали','Details')}</Button></td>
        </tr>)}</tbody></table></div>
      </Panel>
      <div className="st-columns">
        <Panel title={t('Гарантии доставки','Delivery guarantees')}>
          <div className="st-guidance"><div><strong>Outbox → Dispatcher</strong><span>{t('Только подтверждённые события, идемпотентные попытки с backoff и DLQ.','Deliver committed events via outbox, idempotent attempts with backoff and DLQ.')}</span></div><div><strong>HMAC Signature</strong><span>{t('Секрет в Vault; timestamp, nonce и проверка replay на принимающей стороне.','Vault-held secret, timestamp, nonce and replay protection at receivers.')}</span></div></div>
        </Panel>
        <Panel title={t('Безопасность endpoint','Endpoint security')}>
          <div className="st-guidance"><div><strong>HTTPS + SSRF</strong><span>{t('Сервер должен проверять DNS, IP, редиректы и egress policy; клиентская валидация недостаточна.','Server must enforce DNS/IP, redirect and egress protections; client-side validation is insufficient.')}</span></div><div><strong>Access + Audit</strong><span>{t('Смена URL или secretRef требует прав и обязательной записи в Audit.','Endpoint and secret reference changes require authorization and Audit records.')}</span></div></div>
        </Panel>
      </div>
      {webhook?<Modal title={webhook.name} close={()=>setSelected(null)}>
        <div className="st-details"><Detail label="Webhook ID" value={<span className="st-mono">{webhook.id}</span>}/><Detail label="Endpoint" value={<span className="st-mono">{webhook.url}</span>}/><Detail label="Status" value={<Pill tone={webhook.status==='Enabled'?'positive':'warning'}>{webhook.status}</Pill>}/><Detail label="Signing" value={webhook.signing}/><Detail label="Vault SecretRef" value={<span className="st-mono">{webhook.secretRef||'Not configured'}</span>}/><Detail label="Created" value={webhook.createdAt}/></div>
        <Text as="h3" variant="subheader-2">{t('Подписанные события','Subscribed events')}</Text><div className="st-tags">{webhook.events.map(event=><Pill key={event}>{event}</Pill>)}</div>
        <Notice tone="warning">{t('Вебхуки без ссылки на ключ подписи нельзя включить. Симуляция не отправляет POST запросы.','A signing key reference is required to enable webhooks. No POST requests are made in the demo.')}</Notice>
        <div className="st-actions st-actions_end"><Button view="outlined" onClick={()=>{setEditing(webhook.id);setSelected(null)}}>{t('Редактировать','Edit')}</Button><Button view={webhook.status==='Enabled'?'outlined-danger':'action'} onClick={()=>{const ok=toggleWebhook(webhook.id);if(!ok)setNotice(t('Для включения требуется signing secretRef','Signing secretRef required to enable'));else setNotice(t('Демостатус изменён','Demo status updated'))}}>{webhook.status==='Enabled'?t('Отключить','Disable'):t('Включить','Enable')}</Button></div>
      </Modal>:null}
      {(creating||currentEditor)?<WebhookEditor current={currentEditor} key={currentEditor?.id??'new'} t={t} close={()=>{setCreating(false);setEditing(null)}}/>:null}
      {delivery?<WebhookDeliveryDetails key={delivery.id} delivery={delivery} t={t} close={()=>setSelectedDelivery(null)}/>:null}
    </>}
  </SettingsPage>
}

function TokenRequest({close,t}:{close:()=>void;t:T}){
  const [name,setName]=useState('')
  const [scopes,setScopes]=useState<TokenScope[]>(['resources:read'])
  const [expiry,setExpiry]=useState('30')
  const [error,setError]=useState('')
  const toggle=(scope:TokenScope)=>setScopes(prev=>prev.includes(scope)?prev.filter(x=>x!==scope):[...prev,scope])
  return <Modal title={t('Создать заявку на API-токен','Request API token')} close={close}>
    <form className="st-form" onSubmit={e=>{
      e.preventDefault()
      const issue=requestDemoToken({name,scopes,expiryDays:Number(expiry)})
      if(issue)setError(issue);else close()
    }}>
      <Input label={t('Название токена','Token name')} value={name} onChange={setName} placeholder="ci-readonly"/>
      <Select label={t('Срок действия','Expiry')} value={expiry} onChange={setExpiry} options={[7,30,90].map(n=>({value:String(n),label:`${n} days`}))}/>
      <div className="st-field"><span>{t('Разрешения','Requested scopes')}</span><div className="st-checkbox-grid">
        {tokenScopes.map(scope=><label key={scope}><input type="checkbox" checked={scopes.includes(scope)} onChange={()=>toggle(scope)}/><span className="st-mono">{scope}</span></label>)}
      </div></div>
      <Notice tone="warning">{t('Сохраняется только заявка со статусом Draft, действующий ключ НЕ выдаётся. В production выпуск требует M8 Access, политики TTL и безопасного одноразового показа секрета.','Creates a Draft metadata request only; NO usable credential is issued. Production issuance needs M8 Access, TTL policy and secure one-time secret display.')}</Notice>
      {error?<Notice tone="danger">{error}</Notice>:null}
      <div className="st-actions st-actions_end"><Button view="outlined" type="button" onClick={close}>{t('Отмена','Cancel')}</Button><Button view="action" type="submit">{t('Создать черновик','Create draft')}</Button></div>
    </form>
  </Modal>
}
export function SettingsApiTokensPage(){
  const demo=useSettingsDemo()
  const [search,setSearch]=useState('')
  const [status,setStatus]=useState('All')
  const [creating,setCreating]=useState(false)
  const [selected,setSelected]=useState<string|null>(null)
  const [confirm,setConfirm]=useState<string|null>(null)
  const [notice,setNotice]=useState('')
  const visible=useMemo(()=>demo.tokens.filter(token=>
    (status==='All'||token.status===status)&&[token.name,token.id,token.owner,...token.scopes].some(s=>s.toLowerCase().includes(search.toLowerCase()))
  ),[demo.tokens,search,status])
  const token=demo.tokens.find(t=>t.id===selected)
  const tokenToRevoke=demo.tokens.find(t=>t.id===confirm)
  return <SettingsPage tab="api-tokens" title="API Tokens" subtitle="Access scopes, expiry, credential metadata and revocation">
    {t=><>
      <div className="st-section-heading"><div><Text as="h2" variant="header-2">{t('API-токены проекта','Project API tokens')}</Text><Text color="secondary" variant="body-2">{t('Управление метаданными доступов без отображения секретов','Manage access metadata without exposing credentials')}</Text></div>
        <Button view="action" onClick={()=>setCreating(true)}>+ {t('Новый токен','New token request')}</Button>
      </div>
      <div className="st-kpis"><Metric label={t('Токенов','Token records')} value={demo.tokens.length}/><Metric label="Active (sample)" value={demo.tokens.filter(x=>x.status==='Active').length}/><Metric label="Draft" value={demo.tokens.filter(x=>x.status==='Draft').length} tone="warning"/><Metric label="Revoked" value={demo.tokens.filter(x=>x.status==='Revoked').length}/></div>
      <Panel title={t('Реестр API-токенов','API token registry')} subtitle={t('Отпечатки и scopes — только демометаданные. Реальные credentials не существуют.','Fingerprints and scopes are synthetic metadata; no real credentials exist.')}>
        <div className="st-toolbar"><div className="st-search"><Icon data={Magnifier} size={16}/><input aria-label="Search tokens" value={search} onChange={e=>setSearch(e.target.value)} placeholder={t('Имя, владелец, scope','Name, owner, scope')}/></div>
          <Select label="Status" value={status} onChange={setStatus} options={['All','Active','Draft','Revoked'].map(x=>({value:x,label:x}))}/></div>
        <div className="st-table-scroll"><table className="st-table"><thead><tr><th>{t('Токен','Token')}</th><th>{t('Владелец','Owner')}</th><th>Scopes</th><th>Expires</th><th>Last used</th><th>Status</th><th/></tr></thead>
        <tbody>{visible.map(x=><tr key={x.id}><td><strong>{x.name}</strong><span className="st-muted st-mono">{x.fingerprint}</span></td><td>{x.owner}</td><td>{x.scopes.length} scopes</td><td className="st-nowrap">{x.expiresAt}</td><td className="st-nowrap">{x.lastUsed}</td><td><Pill tone={x.status==='Active'?'positive':x.status==='Draft'?'warning':'neutral'}>{x.status}</Pill></td><td><Button size="s" view="outlined" onClick={()=>setSelected(x.id)}>{t('Детали','Details')}</Button></td></tr>)}</tbody></table></div>
        {!visible.length?<Empty>{t('Токены не найдены','No matching token records')}</Empty>:null}
      </Panel>
      {notice?<Notice tone="positive">{notice}</Notice>:null}
      <div className="st-columns">
        <Panel title={t('Принцип минимальных прав','Least privilege')}>
          <div className="st-guidance"><div><strong>Scoped permissions</strong><span>{t('Токен выдаётся только с явно выбранными scopes и минимальным TTL.','Explicit, least-privilege scopes and short TTLs are required.')}</span></div><div><strong>Rotation & revocation</strong><span>{t('Отзыв должен обрабатываться сервером, а не только в UI; события пишутся в Audit.','Revocation is server-authoritative, with Audit events and a rotation workflow.')}</span></div></div>
        </Panel>
        <Panel title={t('Безопасность секретов','Credential security')}>
          <div className="st-guidance"><div><strong>Never show plaintext again</strong><span>{t('В production возможен одноразовый выпуск, хранится только криптостойкий хэш.','Production can issue a credential once and store only a strong token hash.')}</span></div><div><strong>M8 Identity + Access</strong><span>{t('Проверка владельца, разрешений и step-up для чувствительных операций.','Validate owner, scope and step-up for sensitive operations.')}</span></div></div>
        </Panel>
      </div>
      {token?<Modal title={token.name} close={()=>setSelected(null)}>
        <div className="st-details"><Detail label="Token ID" value={<span className="st-mono">{token.id}</span>}/><Detail label="Owner" value={token.owner}/><Detail label="Fingerprint" value={<span className="st-mono">{token.fingerprint}</span>}/><Detail label="Status" value={<Pill tone={token.status==='Active'?'positive':'warning'}>{token.status}</Pill>}/><Detail label="Created" value={token.createdAt}/><Detail label="Expires" value={token.expiresAt}/><Detail label="Last used" value={token.lastUsed}/></div>
        <Text as="h3" variant="subheader-2">{t('Разрешения токена','Token scopes')}</Text><div className="st-tags">{token.scopes.map(scope=><Pill key={scope}>{scope}</Pill>)}</div>
        <Notice tone="warning">{t('Демо-отзыв не меняет реальные IAM-токены, сессии или credentials.','Demo revocation does not change real IAM tokens, sessions or credentials.')}</Notice>
        <div className="st-actions st-actions_end"><Button view="outlined" onClick={()=>setSelected(null)}>{t('Закрыть','Close')}</Button>{token.status!=='Revoked'?<Button view="outlined-danger" onClick={()=>{setConfirm(token.id);setSelected(null)}}><Icon data={Shield} size={16}/>{t('Отозвать (демо)','Revoke (demo)')}</Button>:null}</div>
      </Modal>:null}
      {tokenToRevoke?<Modal title={t('Подтвердить отзыв','Confirm revocation')} close={()=>setConfirm(null)}>
        <Notice tone="warning">{t('Подтвердите локальный отзыв токена','Confirm simulated revocation of token')} <strong>{tokenToRevoke.name}</strong>. {t('Это не отзывает реальные credentials.','Real credentials are not affected.')}</Notice>
        <div className="st-actions st-actions_end"><Button view="outlined" onClick={()=>setConfirm(null)}>{t('Отмена','Cancel')}</Button><Button view="outlined-danger" onClick={()=>{const ok=revokeDemoToken(tokenToRevoke.id);setConfirm(null);if(ok)setNotice(t('Отзыв записан в локальную демомодель','Recorded demo revocation'))}}><Icon data={TriangleExclamation} size={16}/>{t('Отозвать в демо','Revoke in demo')}</Button></div>
      </Modal>:null}
      {creating?<TokenRequest t={t} close={()=>setCreating(false)}/>:null}
    </>}
  </SettingsPage>
}
