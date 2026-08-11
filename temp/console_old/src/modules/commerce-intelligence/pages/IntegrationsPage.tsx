import {Button, Card, Text} from '@gravity-ui/uikit'
import {useQuery} from '@tanstack/react-query'

import {getIntegrations} from '../mock/queries'
import {notifyAction, statusTone} from '../utils'
import {CommercePage, ErrorState, LoadingState, StatusCell} from './pageCommon'

export function IntegrationsPage() {
  const query = useQuery({queryKey: ['commerce-intelligence', 'integrations'], queryFn: getIntegrations})

  return (
    <CommercePage title="Интеграции" subtitle="Подключение источников данных, систем публикации цен и внешних провайдеров.">
      {query.isLoading ? <LoadingState /> : null}
      {query.isError ? <ErrorState onRetry={() => void query.refetch()} /> : null}
      {query.data ? (
        <div className="ci-provider-grid">
          {query.data.integrations.map((provider) => (
            <Card view="outlined" type="container" className="ci-provider" key={provider.provider}>
              <div className="ci-provider__header">
                <div>
                  <Text as="h2" variant="header-1">
                    {provider.provider}
                  </Text>
                  <Text variant="caption-2" color="secondary">
                    Последняя синхронизация: {provider.lastSync}
                  </Text>
                </div>
                <StatusCell value={provider.status} />
              </div>
              <div className="ci-provider__metrics">
                <div>
                  <Text variant="caption-2" color="secondary">Ошибки</Text>
                  <Text variant="header-2">{provider.errors}</Text>
                </div>
                <div>
                  <Text variant="caption-2" color="secondary">Качество данных</Text>
                  <Text variant="header-2">{provider.dataQuality}</Text>
                </div>
                <div>
                  <Text variant="caption-2" color="secondary">Состояние</Text>
                  <span className={`ci-provider__dot ci-provider__dot_${statusTone(provider.status)}`} />
                </div>
              </div>
              <div className="ci-provider__actions">
                {provider.actions.map((action) => (
                  <Button key={action} view={action === 'Отключить' ? 'outlined-danger' : 'outlined'} onClick={() => notifyAction(action, provider.provider)}>
                    {action}
                  </Button>
                ))}
                <Button view="flat" onClick={() => notifyAction('Посмотреть логи', provider.provider)}>
                  Посмотреть логи
                </Button>
              </div>
            </Card>
          ))}
        </div>
      ) : null}
    </CommercePage>
  )
}
