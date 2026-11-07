/* eslint-disable react-refresh/only-export-components */
import type {ReactNode} from 'react'
import {Card, Text} from '@gravity-ui/uikit'
import {Link} from '@tanstack/react-router'

import {ConsoleBreadcrumbs} from '../../components/ConsoleBreadcrumbs'
import {useConsoleI18n} from '../../App'
import type {AlertSeverity,AlertStatus,LogLevel,TraceStatus} from './demo'
import './observability.css'

export type T=(ru:string,en:string)=>string
export type ObservabilityPage='metrics'|'logs'|'traces'|'alerts'|'slo'
export const tabs: {id:ObservabilityPage;path:'/observability/metrics'|'/observability/logs'|'/observability/traces'|'/observability/alerts'|'/observability/slo';ru:string;en:string}[]=[
  {id:'metrics',path:'/observability/metrics',ru:'Метрики',en:'Metrics'},
  {id:'logs',path:'/observability/logs',ru:'Логи',en:'Logs'},
  {id:'traces',path:'/observability/traces',ru:'Трассировки',en:'Traces'},
  {id:'alerts',path:'/observability/alerts',ru:'Алерты',en:'Alerts'},
  {id:'slo',path:'/observability/slo',ru:'SLO',en:'SLO'},
]
export const fmt=(n:number)=>n.toLocaleString('en-US')
export const utc=(value:string)=>value.replace('T',' ').replace(/Z$/,' UTC')
export function Label({children,tone='neutral'}:{children:ReactNode;tone?:'neutral'|'info'|'positive'|'warning'|'danger'}){
  return <span className={`ob-label ob-label_${tone}`}>{children}</span>
}
export function LogBadge({level}:{level:LogLevel}){
  return <Label tone={level==='ERROR'?'danger':level==='WARN'?'warning':level==='DEBUG'?'info':'neutral'}>{level}</Label>
}
export function TraceBadge({status}:{status:TraceStatus}){
  return <Label tone={status==='OK'?'positive':'danger'}>{status}</Label>
}
export function IncidentBadge({status}:{status:AlertStatus}){
  return <Label tone={status==='Firing'?'danger':status==='Acknowledged'?'warning':status==='Resolved'?'positive':'info'}>{status}</Label>
}
export function SeverityBadge({severity}:{severity:AlertSeverity}){
  return <Label tone={severity==='Critical'?'danger':severity==='Warning'?'warning':'neutral'}>{severity}</Label>
}
export function Page({tab,title,subtitle,children}:{tab:ObservabilityPage;title:string;subtitle:string;children:(t:T)=>ReactNode}){
  const {language}=useConsoleI18n()
  const t:T=(ru,en)=>language==='ru'?ru:en
  return <main className="m8-page__body"><section className="m8-page__content ob-page">
    <div className="ob-heading">
      <ConsoleBreadcrumbs items={[{text:'Observability',href:'/observability/metrics'},{text:t(tabs.find(x=>x.id===tab)?.ru??'Метрики',tabs.find(x=>x.id===tab)?.en??'Metrics')}]}/>
      <div className="ob-heading__title"><Text as="h1" variant="display-1">{title}</Text><Label tone="info">DEMO DATA</Label></div>
      <Text as="p" variant="body-2" color="secondary">{subtitle}</Text>
    </div>
    <div className="ob-context"><strong>M8 Platform</strong><span>· Acme / Production · eu-1</span><span className="ob-context__status">Synthetic telemetry</span></div>
    <nav className="ob-tabs" aria-label="Observability navigation">
      {tabs.map(x=><Link key={x.id} to={x.path} className={`ob-tab ${x.id===tab?'ob-tab_active':''}`} aria-current={tab===x.id?'page':undefined}>{t(x.ru,x.en)}</Link>)}
    </nav>
    <div className="ob-demo-note">{t('Демомодель: показатели и события синтетические, действия не изменяют Mimir, Loki, Tempo, Alertmanager и реальные сервисы.','Demo only: telemetry is synthetic; actions do not affect Mimir, Loki, Tempo, Alertmanager or real services.')}</div>
    {children(t)}
  </section></main>
}
export function Panel({title,subtitle,actions,children}:{title:string;subtitle?:string;actions?:ReactNode;children:ReactNode}){
  return <Card view="outlined" className="ob-panel">
    <div className="ob-panel__head"><div><Text as="h2" variant="header-1">{title}</Text>{subtitle?<Text as="p" variant="caption-2" color="secondary">{subtitle}</Text>:null}</div>{actions}</div>
    {children}
  </Card>
}
export function Stat({label,value,note,tone}:{label:string;value:string|number;note?:string;tone?:'danger'|'warning'|'positive'}){
  return <Card view="outlined" className="ob-stat"><Text variant="caption-2" color="secondary">{label}</Text><span className={`ob-stat__value ${tone?'ob-stat__value_'+tone:''}`}>{value}</span>{note?<Text variant="caption-2" color="secondary">{note}</Text>:null}</Card>
}
export function Select({label,value,onChange,options}:{label:string;value:string;onChange:(v:string)=>void;options:Array<{value:string;label:string}>}){
  return <label className="ob-field"><span>{label}</span><select className="ob-select" value={value} onChange={e=>onChange(e.target.value)}>{options.map(o=><option key={o.value} value={o.value}>{o.label}</option>)}</select></label>
}
export function Input({label,value,onChange,placeholder,type='text'}:{label:string;value:string;onChange:(v:string)=>void;placeholder?:string;type?:string}){
  return <label className="ob-field"><span>{label}</span><input className="ob-input" value={value} onChange={e=>onChange(e.target.value)} type={type} placeholder={placeholder}/></label>
}
export function Detail({label,value}:{label:string;value:ReactNode}){
  return <div className="ob-detail"><span>{label}</span><strong>{value}</strong></div>
}
export function Notice({children,tone='info'}:{children:ReactNode;tone?:'info'|'warning'|'danger'|'positive'}){
  return <div className={`ob-notice ob-notice_${tone}`}>{children}</div>
}
export function Empty({children}:{children:ReactNode}){return <div className="ob-empty">{children}</div>}
export function timeLabel(value:string){return value.replace('T',' ').slice(0,19)+' UTC'}
export function Sparkline({values,variant='default',height=68}:{values:number[];variant?:'default'|'danger'|'positive';height?:number}){
  if(values.length===0)return <div className="ob-sparkline" style={{height}}/>
  const min=Math.min(...values),max=Math.max(...values),range=Math.max(.0001,max-min)
  const width=320
  const coords=values.map((v,i)=>[i/(values.length-1||1)*width,8+(1-(v-min)/range)*(height-16)])
  const line=coords.map(([x,y],i)=>`${i?'L':'M'}${x.toFixed(1)},${y.toFixed(1)}`).join(' ')
  const fill=line+` L ${width},${height} L 0,${height} Z`
  return <svg className={`ob-sparkline ob-sparkline_${variant}`} width="100%" height={height} viewBox={`0 0 ${width} ${height}`} preserveAspectRatio="none" role="img" aria-label={`Synthetic trend, ${values.length} points, latest ${values.at(-1)}`}>
    <path className="ob-sparkline__area" d={fill}/>
    <path className="ob-sparkline__stroke" d={line} fill="none" strokeWidth="2.5" strokeLinejoin="round" strokeLinecap="round"/>
  </svg>
}
export function Bar({value,max,variant='default'}:{value:number;max:number;variant?:'default'|'danger'|'positive'}){
  const width=max>0?Math.min(100,Math.max(0,value/max*100)):0
  return <div className="ob-bar" role="meter" aria-valuenow={value} aria-valuemin={0} aria-valuemax={Math.max(max,value)}><span className={`ob-bar__fill ob-bar__fill_${variant}`} style={{width:`${width}%`}}/></div>
}
