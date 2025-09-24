<?php

use Bitrix\Main\Application;
use Bitrix\Main\Config\Option;
use Bitrix\Main\Localization\Loc;

$module_id = 'chelbit.exceltools';

Loc::loadMessages($_SERVER['DOCUMENT_ROOT'] . '/bitrix/modules/main/options.php');
Loc::loadMessages(__FILE__);

if (!$USER->IsAdmin()) {
    $APPLICATION->AuthForm(Loc::getMessage('ACCESS_DENIED'));
    return;
}

$request = Application::getInstance()->getContext()->getRequest();

$tabs = [
    [
        'DIV' => 'edit1',
        'TAB' => Loc::getMessage('MAIN_TAB_SET'),
        'TITLE' => Loc::getMessage('MAIN_TAB_TITLE_SET'),
    ],
];

if ($request->isPost() && check_bitrix_sessid()) {
    $options = [
        'mode', 'service_url', 'cli_path', 'auth_token'
    ];
    foreach ($options as $option) {
        if ($request->getPost($option) !== null) {
            Option::set($module_id, $option, $request->getPost($option));
        }
    }
    CAdminMessage::ShowMessage(['MESSAGE' => Loc::getMessage('OPTIONS_SAVED'), 'TYPE' => 'OK']);
}

$tabControl = new CAdminTabControl('tabControl', $tabs);
?>

<form method="post" action="<?= $APPLICATION->GetCurPage() ?>?mid=<?= htmlspecialcharsbx($module_id) ?>&lang=<?= LANGUAGE_ID ?>">
    <?php
    $tabControl->Begin();
    $tabControl->BeginNextTab();

    $mode = Option::get($module_id, 'mode', 'service');
    $service_url = Option::get($module_id, 'service_url', 'http://127.0.0.1:8080');
    $cli_path = Option::get($module_id, 'cli_path', '');
    $auth_token = Option::get($module_id, 'auth_token', '');
    ?>
    <tr class="heading">
        <td colspan="2"><?= Loc::getMessage('CHELBIT_EXCELTOOLS_GROUP_MAIN') ?></td>
    </tr>
    <tr>
        <td width="40%"><?= Loc::getMessage('CHELBIT_EXCELTOOLS_OPTION_MODE') ?></td>
        <td width="60%">
            <select name="mode">
                <option value="service" <?= ($mode === 'service' ? 'selected' : '') ?>><?= Loc::getMessage('CHELBIT_EXCELTOOLS_MODE_SERVICE') ?></option>
                <option value="cli" <?= ($mode === 'cli' ? 'selected' : '') ?>><?= Loc::getMessage('CHELBIT_EXCELTOOLS_MODE_CLI') ?></option>
            </select>
        </td>
    </tr>
    <tr class="heading">
        <td colspan="2"><?= Loc::getMessage('CHELBIT_EXCELTOOLS_GROUP_SERVICE') ?></td>
    </tr>
    <tr>
        <td width="40%"><?= Loc::getMessage('CHELBIT_EXCELTOOLS_OPTION_SERVICE_URL') ?></td>
        <td width="60%"><input type="text" name="service_url" size="50" value="<?= htmlspecialcharsbx($service_url) ?>"></td>
    </tr>
    <tr>
        <td width="40%"><?= Loc::getMessage('CHELBIT_EXCELTOOLS_OPTION_AUTH_TOKEN') ?></td>
        <td width="60%"><input type="password" name="auth_token" size="50" value="<?= htmlspecialcharsbx($auth_token) ?>"></td>
    </tr>
    <tr class="heading">
        <td colspan="2"><?= Loc::getMessage('CHELBIT_EXCELTOOLS_GROUP_CLI') ?></td>
    </tr>
    <tr>
        <td width="40%"><?= Loc::getMessage('CHELBIT_EXCELTOOLS_OPTION_CLI_PATH') ?></td>
        <td width="60%"><input type="text" name="cli_path" size="50" value="<?= htmlspecialcharsbx($cli_path) ?>"></td>
    </tr>

    <?php
    $tabControl->Buttons();
    ?>
    <input type="submit" name="Update" value="<?= Loc::getMessage('MAIN_SAVE') ?>" title="<?= Loc::getMessage('MAIN_OPT_SAVE_TITLE') ?>" class="adm-btn-save">
    <?= bitrix_sessid_post(); ?>
    <?php $tabControl->End(); ?>
</form>
