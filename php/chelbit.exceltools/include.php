<?php

use Bitrix\Main\Loader;

Loader::registerAutoLoadClasses(
    'chelbit.exceltools',
    [
        // Core
        'Chelbit\\Exceltools\\Client' => 'lib/Client.php',
        'Chelbit\\Exceltools\\Config' => 'lib/Config.php',
        'Chelbit\\Exceltools\\Exporter' => 'lib/Exporter.php',

        // Data Structures
        'Chelbit\\Exceltools\\Data\\Column' => 'lib/Data/Column.php',
        'Chelbit\\Exceltools\\Data\\Schema' => 'lib/Data/Schema.php',
        'Chelbit\\Exceltools\\Data\\Validator' => 'lib/Data/Validator.php',
        'Chelbit\\Exceltools\\Data\\Sheet' => 'lib/Data/Sheet.php',
        'Chelbit\\Exceltools\\Data\\Block' => 'lib/Data/Block.php',
        'Chelbit\\Exceltools\\Data\\Table' => 'lib/Data/Table.php',

        // Enums
        'Chelbit\\Exceltools\\Enums\\Mode' => 'lib/Enums/Mode.php',
        'Chelbit\\Exceltools\\Enums\\Types' => 'lib/Enums/Types.php',

        // Clients
        'Chelbit\\Exceltools\\Http\\HttpClient' => 'lib/Http/HttpClient.php',
        'Chelbit\\Exceltools\\Cli\\CliClient' => 'lib/Cli/CliClient.php',
        'Chelbit\\Exceltools\\Cli\\CliExporter' => 'lib/Cli/CliExporter.php',
    ]
);
