/* eslint-disable react-refresh/only-export-components */
import {useEffect} from 'react'
import type {ReactNode} from 'react'
import {Card,Button,Text} from '@gravity-ui/uikit'
import {Link} from '@tanstack/react-router'

import {ConsoleBreadcrumbs} from '../../components/ConsoleBreadcrumbs'
import {useConsoleI18n} from '../../App'
import './settings.css'

export type T=(ru:string,en:string)=>string
export type SettingsTab='project'|'modules'|'integrations'|'webhooks'|'api-tokens'
export const tabs:{id:SettingsTab;path:'/settings/project'|'/settings/modules'|'/settings/integrations'|'/settings/webhooks'|'/settings/api-tokens';ru:string;en:string}[]=[
  {id:'project',path:'/settings/project',ru:'Проект',en:'Project Settings'},
  {id:'modules',path:'/settings/modules',ru:'Модули',en:'Modules'},
  {id:'integrations',path:'/settings/integrations',ru:'Интеграции',en:'Integrations'},
  {id:'webhooks',path:'/settings/webhooks',ru:'Вебхуки',en:'Webhooks'},
  {id:'api-tokens',path:'/settings/api-tokens',ru:'API-токены',en:'API Tokens'},
]
export const fmt=(value:number)=>value.toLocaleString('en-US')
export function SettingsPage({tab,title,subtitle,children}:{tab:SettingsTab;title:string;subtitle:string;children:(t:T)=>ReactNode}){
  const {language}=useConsoleI18n()
  const t:T=(ru,en)=>language==='ru'?ru:en
  return <main className="m8-page__body"><section className="m8-page__content st-page">
    <div className="st-heading">
      <ConsoleBreadcrumbs items={[{text:'Settings',href:'/settings/project'},{text:t(tabs.find(x=>x.id===tab)?.ru??'Проект',tabs.find(x=>x.id===tab)?.en??'Project')}]}/>
      <div className="st-heading__title"><Text as="h1" variant="display-1">{title}</Text><Pill tone="info">DEMO DATA</Pill></div>
      <Text as="p" color="secondary" variant="body-2">{subtitle}</Text>
    </div>
    <nav className="st-tabs" aria-label="Settings navigation">{tabs.map(item=><Link key={item.id} to={item.path} aria-current={item.id===tab?'page':undefined} className={`st-tab ${item.id===tab?'st-tab_active':''}`}>{t(item.ru,item.en)}</Link>)}</nav>
    <div className="st-context"><strong>Acme / Production / AI Studio</strong><span>{t('Демонстрационные настройки · изменения в памяти браузера, backend не вызывается','Demo settings · changes stay in memory, no backend calls')}</span></div>
    {children(t)}
  </section></main>
}
export function Panel({title,subtitle,actions,children}:{title:string;subtitle?:string;actions?:ReactNode;children:ReactNode}){
  return <Card view="outlined" className="st-panel">
    <div className="st-panel__header"><div><Text as="h2" variant="header-1">{title}</Text>{subtitle?<Text as="p" color="secondary" variant="caption-2">{subtitle}</Text>:null}</div>{actions}</div>
    {children}
  </Card>
}
export function Metric({label,value,caption,tone}:{label:string;value:string|number;caption?:string;tone?:'warning'|'danger'|'positive'}){
  return <Card view="outlined" className="st-metric"><Text variant="caption-2" color="secondary">{label}</Text><span className={`st-metric__value ${tone?'st-metric__value_'+tone:''}`}>{value}</span>{caption?<Text variant="caption-2" color="secondary">{caption}</Text>:null}</Card>
}
export function Pill({children,tone='neutral'}:{children:ReactNode;tone?:'neutral'|'positive'|'warning'|'danger'|'info'}){
  return <span className={`st-pill st-pill_${tone}`}>{children}</span>
}
export function Input({label,value,onChange,type='text',placeholder,disabled}:{label:string;value:string;onChange:(v:string)=>void;type?:string;placeholder?:string;disabled?:boolean}){
  return <label className="st-field"><span>{label}</span><input className="st-input" value={value} onChange={e=>onChange(e.target.value)} type={type} placeholder={placeholder} disabled={disabled}/></label>
}
export function Select({label,value,onChange,options}:{label:string;value:string;onChange:(v:string)=>void;options:Array<{value:string;label:string}>}){
  return <label className="st-field"><span>{label}</span><select className="st-select" value={value} onChange={e=>onChange(e.target.value)}>{options.map(x=><option key={x.value} value={x.value}>{x.label}</option>)}</select></label>
}
export function Textarea({label,value,onChange}:{label:string;value:string;onChange:(v:string)=>void}){
  return <label className="st-field"><span>{label}</span><textarea className="st-textarea" rows={3} value={value} onChange={e=>onChange(e.target.value)}/></label>
}
export function Detail({label,value}:{label:string;value:ReactNode}){return <div className="st-detail"><span>{label}</span><strong>{value}</strong></div>}
export function Notice({children,tone='info'}:{children:ReactNode;tone?:'info'|'warning'|'danger'|'positive'}){return <div className={`st-notice st-notice_${tone}`}>{children}</div>}
export function Empty({children}:{children:ReactNode}){return <div className="st-empty">{children}</div>}
export function Modal({title,close,children}:{title:string;close:()=>void;children:ReactNode}){
  useEffect(()=>{
    const listener=(e:KeyboardEvent)=>{if(e.key==='Escape')close()}
    window.addEventListener('keydown',listener)
    return()=>window.removeEventListener('keydown',listener)
  },[close])
  return <div className="st-modal-host">
    <button type="button" className="st-modal-backdrop" onClick={close} aria-label="Close dialog"/>
    <section className="st-modal" role="dialog" aria-modal="true" aria-label={title}>
      <header className="st-modal__header"><Text as="h2" variant="header-2">{title}</Text><Button view="outlined" onClick={close}>✕</Button></header>
      <div className="st-modal__body">{children}</div>
    </section>
  </div>
}
