<?php

namespace Chelbit\Exceltools\Enums;

class Mode
{
    public const SERVICE = 'service';
    public const CLI = 'cli';

    public static function isValid(string $mode): bool
    {
        return in_array($mode, [self::SERVICE, self::CLI], true);
    }
}
