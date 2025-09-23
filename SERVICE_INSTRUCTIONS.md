# Инструкции по управлению службой (демоном) в Linux

Поскольку ваше приложение теперь может работать как настоящая служба (демон) в Linux, управление и настройка происходят стандартными для системы способами. Ваше название бинарника — `linux_excel_linux`, сервис будет называться `ExcelMS` (как задано в коде).

### 1. Установка и управление службой

Все команды нужно выполнять с правами администратора (`sudo`) из директории, где лежит ваш бинарник.

-   **Установить службу:**
    Эта команда создаст файл юнита для `systemd`.
    ```bash
    sudo ./linux_excel_linux service install
    ```

-   **Запустить службу:**
    ```bash
    sudo ./linux_excel_linux service start
    # или так:
    sudo systemctl start ExcelMS
    ```

-   **Остановить службу:**
    ```bash
    sudo ./linux_excel_linux service stop
    # или так:
    sudo systemctl stop ExcelMS
    ```
-   **Перезапустить службу:**
    ```bash
    sudo ./linux_excel_linux service restart
    # или так:
    sudo systemctl restart ExcelMS
    ```

### 2. Как проверить, что служба запущена

Используйте стандартную команду `systemctl`:

```bash
systemctl status ExcelMS
```

-   Если вы увидите `Active: active (running)` зеленым цветом, значит, все работает.
-   Если `Active: inactive (dead)` или `failed` — служба не запущена или произошла ошибка.

Чтобы посмотреть логи службы (очень полезно для отладки):
```bash
sudo journalctl -u ExcelMS -f
```

### 3. Как задать параметры (PORT, AUTH_TOKEN) для службы

Параметры для системных служб задаются не через аргументы командной строки при запуске, а через специальный **файл окружения**. Это стандартная и безопасная практика в Linux.

Вот как это сделать:

1.  **Создайте файл конфигурации:**
    Например, создадим файл `/etc/excelms.conf`:
    ```bash
    sudo nano /etc/excelms.conf
    ```

2.  **Добавьте в него ваши параметры:**
    Напишите в этом файле переменные окружения, каждая с новой строки:
    ```
    PORT=8090
    AUTH_TOKEN=your_super_secret_token_here
    ```
    Сохраните файл (`Ctrl+X`, затем `Y`).

3.  **Укажите этот файл в настройках службы:**
    *   Откройте для редактирования файл службы, который был создан командой `install`:
        ```bash
        sudo nano /etc/systemd/system/ExcelMS.service
        ```
    *   Найдите секцию `[Service]` и добавьте в нее строку `EnvironmentFile`:
        ```ini
        [Service]
        ...
        EnvironmentFile=/etc/excelms.conf
        ...
        ```
    *   Сохраните файл.

4.  **Перезагрузите конфигурацию systemd и перезапустите службу:**
    Эти команды нужно выполнить, чтобы `systemd` узнал о ваших изменениях.
    ```bash
    sudo systemctl daemon-reload
    sudo systemctl restart ExcelMS
    ```

Теперь ваша служба запустится на порту `8090` и будет использовать указанный токен.

### 4. Удаление службы

Если служба больше не нужна:
```bash
sudo ./linux_excel_linux service stop
sudo ./linux_excel_linux service uninstall
```
