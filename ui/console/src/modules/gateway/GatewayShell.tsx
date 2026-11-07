/* eslint-disable react-refresh/only-export-components */
import {useEffect} from 'react'
import type {ReactNode} from 'react'
import {Cloud, TriangleExclamation} from '@gravity-ui/icons'
import {Button, Card, Icon, Text} from '@gravity-ui/uikit'
import {Link} from '@tanstack/react-router'

import {ConsoleBreadcrumbs} from '../../components/ConsoleBreadcrumbs'
import {useConsoleI18n} from '../../App'
import './gateway.css'

export type T = (ru:string,en:string)=>string
export type GatewayTab = 'services'|'routes'|'consumers'|'policies'|'limits'|'yaml'
export const tabs: Array<{id:GatewayTab;path:'/gateway/services'|'/gateway/routes'|'/gateway/consumers'|'/gateway/policies'|'/gateway/rate-limits'|'/gateway/yaml';ru:string;en:string}> = [
  {id:'services',path:'/gateway/services',ru:'API-сервисы',en:'API Services'},
  {id:'routes',path:'/gateway/routes',ru:'Маршруты',en:'Routes'},
  {id:'consumers',path:'/gateway/consumers',ru:'Потребители',en:'Consumers'},
  {id:'policies',path:'/gateway/policies',ru:'Политики',en:'Policies'},
  {id:'limits',path:'/gateway/rate-limits',ru:'Лимиты запросов',en:'Rate Limits'},
  {id:'yaml',path:'/gateway/yaml',ru:'Gateway YAML',en:'Gateway YAML'},
]
export const formatNumber=(n:number)=>n.toLocaleString('en-US')
export function GatewayPage({active,title,description,actions,children}:{active:GatewayTab;title:string;description:string;actions?:ReactNode;children:(t:T)=>ReactNode}){
  const {language}=useConsoleI18n()
  const t:T=(ru,en)=>language==='ru'?ru:en
  return <main className="m8-page__body"><section className="m8-page__content gw-page">
    <header className="gw-heading">
      <div>
        <ConsoleBreadcrumbs items={[{text:'Gateway',href:'/gateway/services'},{text:t(tabs.find(tab=>tab.id===active)?.ru??'Gateway',tabs.find(tab=>tab.id===active)?.en??'Gateway')}]}/>
        <div className="gw-heading__line"><Text as="h1" variant="display-1">{title}</Text><Pill tone="info">DEMO DATA</Pill></div>
        <Text as="p" variant="body-2" color="secondary">{description}</Text>
      </div>
      {actions?<div className="gw-heading__actions">{actions}</div>:null}
    </header>
    <div className="gw-context">
      <div className="gw-context__cluster"><Icon data={Cloud} size={17}/><strong>gateway-production</strong><span>· Cozystack / eu-1</span></div>
      <span className="gw-context__indicator"/> {t('Конфигурация: Demo Desired State','Configuration: Demo Desired State')}
    </div>
    <nav className="gw-tabs" aria-label="Gateway sections">
      {tabs.map(tab=><Link key={tab.id} className={`gw-tab ${tab.id===active?'gw-tab_active':''}`} to={tab.path} aria-current={tab.id===active?'page':undefined}>{t(tab.ru,tab.en)}</Link>)}
    </nav>
    <div className="gw-demo-banner"><Icon data={TriangleExclamation} size={16}/>{t('Изменения остаются только в памяти браузера. Никаких запросов к Envoy Gateway, Agent Router, Kubernetes или YDB нет.','Changes remain in browser memory only. No requests to Envoy Gateway, Agent Router, Kubernetes or YDB are made.')}</div>
    {children(t)}
  </section></main>
}
export function Panel({title,subtitle,actions,children}:{title:string;subtitle?:string;actions?:ReactNode;children:ReactNode}){
  return <Card view="outlined" className="gw-panel">
    <div className="gw-panel__heading"><div><Text as="h2" variant="header-1">{title}</Text>{subtitle?<Text as="p" variant="caption-2" color="secondary">{subtitle}</Text>:null}</div>{actions}</div>
    {children}
  </Card>
}
export function Metric({label,value,caption,tone}:{label:string;value:string|number;caption?:string;tone?:'warning'|'danger'|'success'}){
  return <Card view="outlined" className="gw-metric"><Text color="secondary" variant="caption-2">{label}</Text><span className={`gw-metric__value ${tone?'gw-metric__value_'+tone:''}`}>{value}</span>{caption?<Text color="secondary" variant="caption-2">{caption}</Text>:null}</Card>
}
export function Pill({children,tone='neutral'}:{children:ReactNode;tone?:'neutral'|'info'|'warning'|'success'|'danger'}){
  return <span className={`gw-pill gw-pill_${tone}`}>{children}</span>
}
export function Status({status}:{status:string}){
  const tone=status==='Healthy'||status==='Accepted'||status==='Active'||status==='Enabled'?'success':status==='Degraded'||status==='Pending'||status==='Unknown'?'warning':status==='Rejected'||status==='Suspended'||status==='Disabled'?'danger':'neutral'
  return <Pill tone={tone}>{status}</Pill>
}
export function Empty({children}:{children:ReactNode}){return <div className="gw-empty">{children}</div>}
export function Detail({label,value}:{label:string;value:ReactNode}){return <div className="gw-detail"><span>{label}</span><strong>{value}</strong></div>}
export function Selector({label,value,options,onChange}:{label:string;value:string;options:Array<{value:string;label:string}>;onChange:(v:string)=>void}){
  return <label className="gw-field"><span>{label}</span><select className="gw-select" value={value} onChange={e=>onChange(e.target.value)}>{options.map(option=><option key={option.value} value={option.value}>{option.label}</option>)}</select></label>
}
export function TextField({label,value,onChange,placeholder,type='text'}:{label:string;value:string;onChange:(v:string)=>void;placeholder?:string;type?:string}){
  return <label className="gw-field"><span>{label}</span><input className="gw-input" type={type} value={value} onChange={e=>onChange(e.target.value)} placeholder={placeholder}/></label>
}
export function DescriptionField({label,value,onChange}:{label:string;value:string;onChange:(v:string)=>void}){
  return <label className="gw-field"><span>{label}</span><textarea className="gw-textarea" rows={3} value={value} onChange={e=>onChange(e.target.value)}/></label>
}
export function Drawer({title,subtitle,close,children}:{title:string;subtitle?:string;close:()=>void;children:ReactNode}){
  useEffect(()=>{
    const listener=(event:KeyboardEvent)=>{if(event.key==='Escape')close()}
    window.addEventListener('keydown',listener)
    return()=>window.removeEventListener('keydown',listener)
  },[close])
  return <div className="gw-drawer-host">
    <button type="button" className="gw-drawer-backdrop" aria-label="Close panel" onClick={close}/>
    <aside className="gw-drawer" role="dialog" aria-modal="true" aria-label={title}>
      <header className="gw-drawer__head"><div><Text as="h2" variant="header-2">{title}</Text>{subtitle?<Text as="p" variant="caption-2" color="secondary">{subtitle}</Text>:null}</div><Button view="outlined" onClick={close} aria-label="Close panel">✕</Button></header>
      <div className="gw-drawer__body">{children}</div>
    </aside>
  </div>
}
export function Notice({children,kind='warning'}:{children:ReactNode;kind?:'warning'|'danger'|'info'|'success'}){
  return <div role={kind==='danger'?'alert':undefined} className={`gw-notice gw-notice_${kind}`}>{children}</div>
}
export function UsageBar({value,total}:{value:number;total:number}){
  const percent=total>0?Math.max(0,Math.min(100,value/total*100)):0
  return <div className="gw-usage" role="meter" aria-valuenow={value} aria-valuemin={0} aria-valuemax={Math.max(total,value)}><span style={{width:`${percent}%`}}/></div>
}
