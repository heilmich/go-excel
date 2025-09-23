<?php

use Bitrix\Main\Loader;

Loader::registerAutoLoadClasses(
    'chelbit.exceltools',
    [
        'Chelbit\\Exceltools\\Client' => 'lib/Client.php',
        'Chelbit\\Exceltools\\CliClient' => 'lib/CliClient.php',
        'Chelbit\\Exceltools\\CliExporter' => 'lib/CliExporter.php',
        'Chelbit\\Exceltools\\HttpClient' => 'lib/HttpClient.php',
        'Chelbit\\Exceltools\\Config' => 'lib/Config.php',
        'Chelbit\\Exceltools\\Column' => 'lib/Column.php',
        'Chelbit\\Exceltools\\Schema' => 'lib/Schema.php',
        'Chelbit\\Exceltools\\Types' => 'lib/Types.php',
        'Chelbit\\Exceltools\\Exporter' => 'lib/Exporter.php',
        'Chelbit\\Exceltools\\Validator' => 'lib/Validator.php',
    ]
);
