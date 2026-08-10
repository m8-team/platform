import type {Integration} from './types'

export const integrations: Integration[] = [
  {provider: '1C Adapter', status: 'Норма', lastSync: '5 мин назад', errors: 0, dataQuality: '98.1%', actions: ['Настроить', 'Проверить подключение', 'Запустить синхронизацию']},
  {provider: 'Competitor Data Provider', status: 'Норма', lastSync: '12 мин назад', errors: 2, dataQuality: '96.3%', actions: ['Настроить', 'Посмотреть логи']},
  {provider: 'Metacommerce', status: 'Риск', lastSync: '38 мин назад', errors: 7, dataQuality: '91.4%', actions: ['Проверить подключение', 'Посмотреть логи']},
  {provider: 'Priceva', status: 'Норма', lastSync: '21 мин назад', errors: 0, dataQuality: '97.8%', actions: ['Настроить', 'Запустить синхронизацию']},
  {provider: 'uXprice', status: 'Норма', lastSync: '18 мин назад', errors: 1, dataQuality: '95.9%', actions: ['Проверить подключение', 'Посмотреть логи']},
  {provider: 'ALL RIVAL', status: 'Исключение', lastSync: '2 ч назад', errors: 14, dataQuality: '88.2%', actions: ['Настроить', 'Посмотреть логи', 'Отключить']},
  {provider: 'GoodsForecast', status: 'Норма', lastSync: '9 мин назад', errors: 0, dataQuality: '99.0%', actions: ['Настроить', 'Проверить подключение']},
  {provider: 'Loginom', status: 'Норма', lastSync: '16 мин назад', errors: 0, dataQuality: '97.1%', actions: ['Настроить', 'Запустить синхронизацию']},
  {provider: 'S3 / FTP Import', status: 'Риск', lastSync: '1 ч назад', errors: 4, dataQuality: '93.5%', actions: ['Проверить подключение', 'Посмотреть логи']},
  {provider: 'Kafka / YDB Topics', status: 'Норма', lastSync: 'в реальном времени', errors: 0, dataQuality: '98.7%', actions: ['Настроить', 'Посмотреть логи']},
  {provider: 'ClickHouse', status: 'Норма', lastSync: '3 мин назад', errors: 0, dataQuality: '99.2%', actions: ['Проверить подключение', 'Посмотреть логи']},
  {provider: 'MLflow', status: 'Норма', lastSync: '25 мин назад', errors: 1, dataQuality: '96.8%', actions: ['Настроить', 'Запустить синхронизацию']},
]
