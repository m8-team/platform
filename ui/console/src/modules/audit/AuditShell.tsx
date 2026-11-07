import type {ReactNode} from 'react'
import {Card, Text} from '@gravity-ui/uikit'
import {Link} from '@tanstack/react-router'
import {ConsoleBreadcrumbs} from '../../components/ConsoleBreadcrumbs'
import {useConsoleI18n} from '../../App'
import type {AuditSeverity, AuditOutcome} from './demo'
import './audit.css'

export type T=(ru:string,en:string)=>string
export const fmt=(n:number)=>n.toLocaleString('en-US')
export const auditEventPath='/audit/events'
export const auditExportPath='/audit/exports'
export const dayUTC=(timestamp:string)=>timestamp.slice(0,10)
export const dateTimeUTC=(timestamp:string)=>timestamp.replace('T',' ').replace(/Z$/,' UTC')

export function Chip({children,tone='neutral'}:{children:ReactNode;tone?:'neutral'|'positive'|'warning'|'danger'|'info'}){
  return <span className={`au-chip au-chip_${tone}`}>{children}</span>
}
export function Outcome({outcome}:{outcome:AuditOutcome}){
  return <Chip tone={outcome==='Success'?'positive':outcome==='Denied'?'warning':'danger'}>{outcome}</Chip>
}
export function Severity({severity}:{severity:AuditSeverity}){
  return <Chip tone={severity==='Critical'?'danger':severity==='Warning'?'warning':'neutral'}>{severity}</Chip>
}
export function AuditLayout({page,details,children}:{page:'events'|'exports';details?:string;children:(t:T)=>ReactNode}){
  const {language}=useConsoleI18n()
  const t:T=(ru,en)=>language==='ru'?ru:en
  return <main className="m8-page__body"><section className="m8-page__content au-page">
    <div className="au-heading">
      <ConsoleBreadcrumbs items={[
        {text:'Audit',href:auditEventPath},
        ...(details?[{text:t('События','Events'),href:auditEventPath}]:[]),
        {text:details??(page==='events'?t('События','Events'):t('Экспорт','Exports'))},
      ]}/>
      <div className="au-heading__title"><Text as="h1" variant="display-1">{details?t('Событие аудита','Audit Event'):page==='events'?t('Журнал аудита','Audit Events'):t('Экспорт аудита','Audit Exports')}</Text><Chip tone="info">DEMO DATA</Chip></div>
      <Text as="p" variant="body-2" color="secondary">{details?details:t('Аудируемые действия в M8 Platform, контроль изменений и экспорт для расследований','Security and operational activity in M8 Platform, change tracking and evidence exports')}</Text>
    </div>
    <nav className="au-tabs" aria-label="Audit navigation">
      <Link to={auditEventPath} className={`au-tab ${page==='events'?'au-tab_active':''}`} aria-current={page==='events'?'page':undefined}>{t('События','Events')}</Link>
      <Link to={auditExportPath} className={`au-tab ${page==='exports'?'au-tab_active':''}`} aria-current={page==='exports'?'page':undefined}>{t('Экспорт','Exports')}</Link>
    </nav>
    <div className="au-demo-banner">{t('Только демоданные · события сохраняются неизменными в рамках вкладки; нет подключения к YDB, audit ledger или внешним хранилищам.','Demo data only · events are append-only within this tab; no connection to YDB, an audit ledger or external storage.')}</div>
    {children(t)}
  </section></main>
}
export function Panel({title,subtitle,actions,children}:{title:string;subtitle?:string;actions?:ReactNode;children:ReactNode}){
  return <Card view="outlined" className="au-panel"><div className="au-panel__head"><div><Text as="h2" variant="header-1">{title}</Text>{subtitle?<Text as="p" variant="caption-2" color="secondary">{subtitle}</Text>:null}</div>{actions}</div>{children}</Card>
}
export function Stat({label,value,caption,tone}:{label:string;value:string|number;caption?:string;tone?:'warning'|'danger'}){
  return <Card view="outlined" className="au-stat"><Text color="secondary" variant="caption-2">{label}</Text><span className={`au-stat__value ${tone?'au-stat__value_'+tone:''}`}>{value}</span>{caption?<Text color="secondary" variant="caption-2">{caption}</Text>:null}</Card>
}
export function SelectFilter({label,value,onChange,options}:{label:string;value:string;onChange:(next:string)=>void;options:Array<{value:string;label:string}>}){
  return <label className="au-filter"><span>{label}</span><select className="au-select" value={value} onChange={e=>onChange(e.target.value)}>{options.map(option=><option key={option.value} value={option.value}>{option.label}</option>)}</select></label>
}
export function TextField({label,value,onChange,placeholder,type='text'}:{label:string;value:string;onChange:(next:string)=>void;placeholder?:string;type?:string}){
  return <label className="au-filter"><span>{label}</span><input className="au-input" type={type} value={value} placeholder={placeholder} onChange={e=>onChange(e.target.value)}/></label>
}
export function Field({label,children}:{label:string;children:ReactNode}){
  return <div className="au-field"><Text variant="caption-2" color="secondary">{label}</Text><div>{children}</div></div>
}
export function Empty({children}:{children:ReactNode}){return <div className="au-empty">{children}</div>}
export function Notice({children,type='info'}:{children:ReactNode;type?:'info'|'warning'|'danger'|'success'}){return <div className={`au-notice au-notice_${type}`}>{children}</div>}
export function jsonText(value:unknown){return JSON.stringify(value,null,2)}
