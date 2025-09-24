<?php

use Bitrix\Main\Localization\Loc;
use Bitrix\Main\ModuleManager;

Loc::loadMessages(__FILE__);

class chelbit_exceltools extends CModule
{
    public $MODULE_ID = 'chelbit.exceltools';
    public $MODULE_VERSION;
    public $MODULE_VERSION_DATE;
    public $MODULE_NAME;
    public $MODULE_DESCRIPTION;
    public $PARTNER_NAME;
    public $PARTNER_URI;

    public function __construct()
    {
        $arModuleVersion = [];
        include(__DIR__ . '/version.php');
        $this->MODULE_VERSION = $arModuleVersion['VERSION'];
        $this->MODULE_VERSION_DATE = $arModuleVersion['VERSION_DATE'];
        $this->MODULE_NAME = Loc::getMessage('CHELBIT_EXCELTOOLS_MODULE_NAME');
        $this->MODULE_DESCRIPTION = Loc::getMessage('CHELBIT_EXCELTOOLS_MODULE_DESC');
        $this->PARTNER_NAME = Loc::getMessage('CHELBIT_EXCELTOOLS_PARTNER_NAME');
        $this->PARTNER_URI = Loc::getMessage('CHELBIT_EXCELTOOLS_PARTNER_URI');
    }

    public function DoInstall()
    {
        ModuleManager::registerModule($this->MODULE_ID);
        // We can add DB tables, event handlers etc. here if needed in the future
    }

    public function DoUninstall()
    {
        // Clean up options
        COption::RemoveOption($this->MODULE_ID);
        ModuleManager::unRegisterModule($this->MODULE_ID);
    }
}
