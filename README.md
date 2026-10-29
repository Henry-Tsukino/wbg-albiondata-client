# WBG Albion Online Data Client

[![GitHub Release](https://img.shields.io/github/v/tag/Henry-Tsukino/wbg-albiondata-client?label=версия)](https://github.com/Henry-Tsukino/wbg-albiondata-client/releases) [![Downloads](https://img.shields.io/github/downloads/Henry-Tsukino/wbg-albiondata-client/total?label=установок)](https://github.com/Henry-Tsukino/wbg-albiondata-client/releases) [![Go](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white)](https://go.dev) [![Windows](https://img.shields.io/badge/Windows-0078D4?logo=windows&logoColor=white)](https://github.com/Henry-Tsukino/wbg-albiondata-client/blob/main/Windows.md#установка-клиента-для-windows) [![Linux](https://img.shields.io/badge/Linux-FCC624?logo=linux&logoColor=black)](https://github.com/Henry-Tsukino/wbg-albiondata-client/blob/main/Linux.md#установка-клиента-для-linux)

Клиент мониторинга сетевого трафика, сделанный для [Albion Online Data Project](https://www.albion-online-data.com/).

**Это форк [ao-data/albiondata-client](https://github.com/ao-data/albiondata-client). Для просмотра и использования официального проекта и его кода, пожалуйста, посетите оригинальный репозиторий.**

**[Детальное описание работы проекта](https://github.com/Henry-Tsukino/wbg-albiondata-client?tab=readme-ov-file#описание-проекта) находится сразу под разделом работы с клиентом(установка/настройка/удаление)**

## Работа с клиентом

### Инсталляция клиента

Скачайте последний релиз клиента под вашу систему из: https://github.com/Henry-Tsukino/wbg-albiondata-client/releases

В зависимости от вашей системы, воспользуйтесь следующими инструкциями:

- [Windows 10/11: инструкция по инсталляции клиента](https://github.com/Henry-Tsukino/wbg-albiondata-client/blob/main/Windows.md#установка-клиента-для-windows)
  - [Инструкция по настройке](https://github.com/Henry-Tsukino/wbg-albiondata-client/blob/main/Windows.md#настройка-клиента)
  - [Инструкция по деинсталляции](https://github.com/Henry-Tsukino/wbg-albiondata-client/blob/main/Windows.md#удаление-клиента)
- [Linux: инструкция по инсталляции клиента](https://github.com/Henry-Tsukino/wbg-albiondata-client/blob/main/Linux.md#установка-клиента-для-linux)
- [MacOS: **Временно не поддерживается**(в разработке)](https://github.com/Henry-Tsukino/wbg-albiondata-client/blob/main/MacOS.md)

### Поиск и исправление неисправностей

**В случае, если клиент после запуска:**

* Не появился в системном трее
* Не запускается
* Выдает ошибки
* Ведет себя как-то странно
* Использует слишком много ресурсов вашего устройства
* Не появилось пинга в дискорде
* Программа по вашему мнению работает неисправно

**Сделайте следующие действия:**

1. Зайдите в папку в которую вы установили клиент
   * Кликните ПКМ на ярлык клиента и после выберите "Открыть расположение файла" в открывшемся меню
2. Найдите в файлах клиента файл **albiondata-client.log**
3. Скиньте данный файл одному из разработчиков, указанных ниже, связавшись с ним через Дискорд:
   * [![Discord](https://img.shields.io/badge/Discord-pinok.tsukino-f0c5cd?logo=discord&logoColor=white)](https://discordapp.com/users/747787435198513153/ "Discord Pinok")
   * [![Discord](https://img.shields.io/badge/Discord-henry.tsukino-ff8243?logo=discord&logoColor=white)](https://discordapp.com/users/603673234138988591/ "Discord Henry Tsukino")

## Описание проекта

### Суть работы

Albion Data Client мониторит локальный сетевой трафик вашего устройства, идентифицирует UDP пакеты, содержащие данные игры Albion Online и после отправляет информацию на центральый сервер брокера сообщений NATS для дальнейшего распределения по другим сервисам. Таким образом получается сбор данных, осуществляемый силами сообщества, для аналитики рынка, слежения за событиями и мониторинга игровой статистики.

### Функции

- Мониторинг сетевого трафика в реальном времени с помощью libpcap
- Фильтрование UDP пакетов и их дальнейший анализ для поиска протоколов Albion Online
- Интеграция с системой брокера сообщений NATS
- Поддержка кросс-платформы (Windows, Linux, ~~macOS~~) (macOS в процессе реализации, пока сделано только у оригинала)
- Интеграция в системный трей Windows ~~и macOS~~
- Минимальное потребление ресурсов системы:
  - Максимум потребляет около [20мб ОЗУ](https://github.com/Henry-Tsukino/wbg-albiondata-client/blob/main/image/README/20mb_RAM.png)[^1] и 25мб памяти накопителя

### Карается ли баном?

Заявление от **SBI Games**(разработчика) касательно мониторинга сетевых пакетов:

> Our position is quite simple. As long as you just look and analyze we are ok with it. The moment you modify or manipulate something or somehow interfere with our services we will react (e.g. perma-ban, take legal action, whatever).
>
> — MadDave, Technical Lead for Albion Online

Перевод:

> Наша позиция довольно проста. Пока вы просто наблюдаете и анализируете, нас это устраивает. Как только вы что-то измените или манипулируете, или каким-либо образом вмешаетесь в работу наших сервисов, мы отреагируем (например, перманентная блокировка, юридические действия и т. д.).
>
> — MadDave, Технический Руководитель Albion Online

Полный оригинал: https://forum.albiononline.com/index.php/Thread/51604-Is-it-allowed-to-scan-your-internet-trafic-and-pick-up-logs/?postID=512670#post512670

Пользователи должны следовать [Договору об условиях предоставления услуг Albion Online](https://albiononline.com/terms_and_conditions), а также [Правилам игры AlbionOnline](https://albiononline.com/game-rules), при использовании данной программы.

## Сообщество и поддержка

Данный клиент был в первую очередь создан для нужд гильдии World Bank Group.
По интерисующим вас вопросам можете связаться через следующие контакты:

* Дискорд ГИ: [![Discord](https://img.shields.io/badge/Discord-WBG-e29f04?logo=discord&logoColor=white)](https://discord.gg/kgXREFJzsS "Discord World Bank Group guild")
* Дискорд разработчиков форка:
  * [![Discord](https://img.shields.io/badge/Discord-pinok.tsukino-f0c5cd?logo=discord&logoColor=white)](https://discordapp.com/users/747787435198513153/ "Discord Pinok")
  * [![Discord](https://img.shields.io/badge/Discord-henry.tsukino-ff8243?logo=discord&logoColor=white)](https://discordapp.com/users/603673234138988591/ "Discord Henry Tsukino")

## Лицензия

Данный форк использует такую же лицензию "[MIT License](https://github.com/Henry-Tsukino/wbg-albiondata-client/blob/main/LICENSE "MIT License info")", как и оригинальный проект.

[^1]: При выходе обновлений Альбиона, изменения в нем иногда, хотя и крайне редко, могут сломать программу и она начнет активно занимать оперативу на вашем ПК. Закройте программу и сообщите об этом разработчикам форка.
