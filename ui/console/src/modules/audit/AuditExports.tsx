import {useMemo, useState} from 'react'
import {ArrowRotateRight, Check, Clock, Code} from '@gravity-ui/icons'
import {Button, Icon, Text} from '@gravity-ui/uikit'
import {Link} from '@tanstack/react-router'

import {createAuditExport,getExportContent,listAuditEvents,prepareAuditExport,useAuditDemo} from './demo'
import type {AuditCategory,AuditExport,AuditFilters,AuditOutcome,AuditScope,AuditSeverity,ExportFormat} from './demo'
import {AuditLayout,Chip,Empty,Notice,Panel,SelectFilter,Stat,TextField,dateTimeUTC,fmt} from './AuditShell'
import type {T} from './AuditShell'

const categories:AuditCategory[]=['Resource','Gateway','Quota','Identity','Security','System']
const outcomes:AuditOutcome[]=['Success','Failure','Denied']
const severities:AuditSeverity[]=['Info','Warning','Critical']
const scopes:AuditScope[]=['Organization','Project']
const options=(all:string,values:string[])=>[{value:'All',label:all},...values.map(value=>({value,label:value}))]

function exportSizeHint(exportItem:AuditExport):string{
  return `${fmt(exportItem.rowCount)} ${exportItem.rowCount===1?'event':'events'}`
}

function ExportHistory({job,t}:{job:AuditExport;t:T}){
  const [downloadError,setDownloadError]=useState('')
  const download=()=>{
    const content=getExportContent(job.id)
    if(content===null){setDownloadError(t('Экспорт ещё не подготовлен','Export is not ready'));return}
    try{
      const type=job.format==='CSV'?'text/csv;charset=utf-8':'application/x-ndjson;charset=utf-8'
      const blob=new Blob([content],{type})
      const url=URL.createObjectURL(blob)
      const anchor=document.createElement('a')
      anchor.href=url
      anchor.download=job.filename
      document.body.appendChild(anchor)
      anchor.click()
      anchor.remove()
      window.setTimeout(()=>URL.revokeObjectURL(url),1000)
      setDownloadError('')
    }catch{
      setDownloadError(t('Не удалось создать локальный файл','Could not create local download'))
    }
  }
  return <div className="au-export-row">
    <div className="au-export-row__main">
      <strong className="au-mono">{job.id}</strong>
      <span className="au-muted">{dateTimeUTC(job.createdAt)} · {job.requestedBy}</span>
      <span className="au-muted">{job.filename} · {exportSizeHint(job)} · {t('Безопасная редакция','Safe redaction')}</span>
    </div>
    <Chip tone={job.status==='Ready'?'positive':'warning'}>{job.status}</Chip>
    <div className="au-actions">
      {job.status==='Queued'?<Button size="s" view="outlined" onClick={()=>prepareAuditExport(job.id)}><Icon data={Clock} size={15}/>{t('Подготовить (демо)','Prepare (demo)')}</Button>:
        <Button size="s" view="action" onClick={download}><Icon data={Code} size={15}/>{t('Скачать','Download')}</Button>}
    </div>
    {downloadError?<Notice type="danger">{downloadError}</Notice>:null}
  </div>
}

export function AuditExportsPage(){
  const demo=useAuditDemo()
  const [filters,setFilters]=useState<AuditFilters>(demo.draft)
  const [format,setFormat]=useState<ExportFormat>('CSV')
  const [error,setError]=useState('')
  const [showPreview,setShowPreview]=useState(false)
  const invalidRange=Boolean(filters.fromDate&&filters.toDate&&filters.fromDate>filters.toDate)
  const candidates=useMemo(()=>invalidRange?[]:listAuditEvents(filters,demo.events),[filters,demo.events,invalidRange])
  const update=(key:keyof AuditFilters,value:string)=>{
    setFilters(prev=>({...prev,[key]:value==='All'?'':value}))
    setShowPreview(false)
    setError('')
  }
  const submit=()=>{
    if(invalidRange){setError('Date range is invalid');return}
    createAuditExport(filters,format)
    setShowPreview(false)
    setError('')
  }
  return <AuditLayout page="exports">{t=><>
    <div className="au-section-title">
      <div><Text as="h2" variant="header-2">{t('Экспорт данных аудита','Audit data exports')}</Text><Text color="secondary" variant="body-2">{t('Создание выборки, подготовка и локальная выгрузка в CSV/JSONL','Create a snapshot, prepare and download a local CSV or JSONL file')}</Text></div>
      <Link to="/audit/events" className="au-link">{t('Открыть журнал событий','Browse events')}</Link>
    </div>
    <div className="au-kpis">
      <Stat label={t('Запросов на экспорт','Export requests')} value={demo.exports.length}/>
      <Stat label={t('Ожидают подготовки','Queued')} value={demo.exports.filter(job=>job.status==='Queued').length}/>
      <Stat label={t('Готовы','Ready')} value={demo.exports.filter(job=>job.status==='Ready').length}/>
      <Stat label={t('Событий в журнале','Journal events')} value={fmt(demo.events.length)} caption={t('До фильтрации','Before filters')}/>
    </div>
    <div className="au-columns au-columns_wide">
      <Panel title={t('Новая выгрузка','Create export')} subtitle={t('Сначала фиксируется выборка на момент создания заявки; затем файл подготавливается отдельно','Selection is snapshotted at request time, then prepared separately')}>
        <form className="au-export-form" onSubmit={event=>{event.preventDefault();submit()}}>
          <div className="au-form-grid">
            <SelectFilter label={t('Формат','Format')} value={format} onChange={v=>setFormat(v as ExportFormat)} options={[{value:'CSV',label:'CSV'},{value:'JSONL',label:'JSON Lines'}]}/>
            <SelectFilter label={t('Категория','Category')} value={filters.category||'All'} onChange={v=>update('category',v)} options={options(t('Все','All'),categories)}/>
            <SelectFilter label={t('Результат','Outcome')} value={filters.outcome||'All'} onChange={v=>update('outcome',v)} options={options(t('Все','All'),outcomes)}/>
            <SelectFilter label={t('Важность','Severity')} value={filters.severity||'All'} onChange={v=>update('severity',v)} options={options(t('Все','All'),severities)}/>
            <SelectFilter label="Scope" value={filters.scope||'All'} onChange={v=>update('scope',v)} options={options(t('Все','All'),scopes)}/>
            <TextField label={t('С даты (UTC)','From (UTC)')} type="date" value={filters.fromDate??''} onChange={v=>update('fromDate',v)}/>
            <TextField label={t('По дату (UTC)','To (UTC)')} type="date" value={filters.toDate??''} onChange={v=>update('toDate',v)}/>
            <TextField label={t('Поиск','Search')} value={filters.query??''} onChange={v=>update('query',v)} placeholder="actor / operation / resource"/>
          </div>
          {invalidRange?<Notice type="danger">{t('Дата начала позже даты окончания','Start date is after end date')}</Notice>:null}
          {error?<Notice type="danger">{error}</Notice>:null}
          <div className="au-export-summary">
            <div><span>{t('Оценка объёма','Matching events')}</span><strong>{fmt(candidates.length)}</strong></div>
            <div><span>{t('Редактирование полей','Redaction')}</span><Chip tone="positive">Safe</Chip></div>
            <div><span>{t('Подтверждение целостности','Integrity verification')}</span><Chip tone="warning">Not verified</Chip></div>
          </div>
          <div className="au-actions au-actions_end">
            <Button type="button" view="outlined" onClick={()=>setShowPreview(v=>!v)}>{showPreview?t('Скрыть выборку','Hide preview'):t('Предпросмотр','Preview')}</Button>
            <Button type="submit" view="action" disabled={invalidRange}>{t('Создать заявку на экспорт','Request export')}</Button>
          </div>
          {showPreview?<div className="au-preview">
            <strong>{t('Первые записи выборки','Sample matching events')}</strong>
            {candidates.slice(0,5).map(event=><div key={event.id}><span className="au-mono">{event.id}</span> · {event.action} · {event.actor.displayName}</div>)}
            {!candidates.length?<Empty>{t('В выборке нет событий','No events in selected range')}</Empty>:null}
          </div>:null}
        </form>
      </Panel>
      <Panel title={t('Безопасность экспорта','Export safeguards')} subtitle={t('Демо-правила и требования к production','Demo behavior and production requirements')}>
        <div className="au-guidance">
          <div><Icon data={Check} size={18}/><span>{t('Экспорт использует снимок выбранных event IDs, новые события в него не добавляются','Export snapshots selected event IDs, later events do not change the selection')}</span></div>
          <div><Icon data={Check} size={18}/><span>{t('Поля с паролями, токенами, ключами и секретами скрываются','Password, token, API key, credential and secret fields are redacted')}</span></div>
          <div><Icon data={Clock} size={18}/><span>{t('Queued → Ready моделируется вручную: фоновые задачи не запускаются','Queued → Ready is advanced manually; no background job runs')}</span></div>
          <div><Icon data={ArrowRotateRight} size={18}/><span>{t('Запрос и подготовка экспорта порождают новые события аудита','Export request and preparation append new audit events')}</span></div>
        </div>
        <Notice type="warning">{t('На практике требуется ограничение прав M8 Access, серверный экспорт, согласование, срок хранения выгрузок и аудит скачивания. Браузерный файл не является доказательством неизменности.','Production requires M8 Access permissions, server-side export, approval, TTL, and download audit. Browser-created files are not tamper-evident evidence.')}</Notice>
      </Panel>
    </div>
    <Panel title={t('История выгрузок','Export history')} subtitle={t('Созданные в этой вкладке заявки; без серверного хранилища','Requests made in this browser tab, no server storage')}>
      {demo.exports.length?<div className="au-export-list">{demo.exports.map(job=><ExportHistory key={job.id} job={job} t={t}/>)}</div>:<Empty>{t('Пока нет выгрузок. Выберите параметры и создайте первую заявку.','No exports yet. Select filters and create an export request.')}</Empty>}
    </Panel>
    <Panel title={t('Предлагаемый жизненный цикл','Proposed production lifecycle')}>
      <div className="au-export-flow">
        <div><span>01</span><strong>Requested</strong><small>{t('Записать событие аудита','Append audit event')}</small></div>
        <div><span>02</span><strong>Authorized</strong><small>{t('RBAC и privacy policy','RBAC and privacy policy')}</small></div>
        <div><span>03</span><strong>Prepared</strong><small>{t('Серверный снимок, redaction','Server snapshot + redaction')}</small></div>
        <div><span>04</span><strong>Ready</strong><small>{t('Подписанный manifest/checksum','Signed manifest / checksum')}</small></div>
        <div><span>05</span><strong>Expired</strong><small>{t('Удаление объекта по TTL','Object TTL / cleanup')}</small></div>
      </div>
      <Notice type="info">{t('В демо реализованы только Requested, Queued, Ready и локальное скачивание.','Only Requested, Queued, Ready and local downloads are implemented in this demo.')}</Notice>
    </Panel>
  </>}</AuditLayout>
}
