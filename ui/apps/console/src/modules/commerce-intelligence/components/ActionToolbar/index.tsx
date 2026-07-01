import {Button} from '@gravity-ui/uikit'

export function ActionToolbar({
  selectedCount,
  onApprove,
  onReject,
  onSchedule,
  onMore,
}: {
  selectedCount: number
  onApprove: () => void
  onReject: () => void
  onSchedule: () => void
  onMore?: () => void
}) {
  return (
    <div className="ci-action-toolbar">
      <span>{selectedCount} выбрано</span>
      <Button view="action" disabled={selectedCount === 0} onClick={onApprove}>
        Согласовать
      </Button>
      <Button view="outlined" disabled={selectedCount === 0} onClick={onReject}>
        Отклонить
      </Button>
      <Button view="outlined" disabled={selectedCount === 0} onClick={onSchedule}>
        Запланировать
      </Button>
      <Button view="flat" onClick={onMore}>
        Еще действия
      </Button>
    </div>
  )
}
