<?php

namespace Chelbit\Exceltools;

use Bitrix\Main\Config\Option;

class Config
{
    private const MODULE_ID = 'chelbit.exceltools';

    public static function getMode(): string
    {
        return Option::get(self::MODULE_ID, 'mode', 'service');
    }

    public static function getServiceUrl(): string
    {
        return Option::get(self::MODULE_ID, 'service_url', 'http://127.0.0.1:8080');
    }

    public static function getCliPath(): string
    {
        return Option::get(self::MODULE_ID, 'cli_path', '');
    }

    public static function getAuthToken(): string
    {
        return Option::get(self::MODULE_ID, 'auth_token', '');
    }
}
