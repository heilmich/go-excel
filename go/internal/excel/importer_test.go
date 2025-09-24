package excel

import (
	"bytes"
	"testing"

	"chelbit/excelms/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func createTestExcelFileForImport(t *testing.T) *bytes.Buffer {
	f := excelize.NewFile()
	sheetName := "UserData"
	_ = f.SetSheetName("Sheet1", sheetName)

	_ = f.SetCellValue(sheetName, "A1", "John Doe")
	_ = f.SetCellValue(sheetName, "B1", "john@example.com")
	_ = f.SetCellValue(sheetName, "D5", "Admin")
	_ = f.SetCellValue(sheetName, "E5", "5")
	_ = f.SetCellValue(sheetName, "D6", "User")
	_ = f.SetCellValue(sheetName, "E6", "2")

	buf, err := f.WriteToBuffer()
	require.NoError(t, err)
	return buf
}

func TestImporter_Process_Valid(t *testing.T) {
	excelFileBytes := createTestExcelFileForImport(t)

	request := &models.Request{
		Sheets: []models.Sheet{
			{
				Name: "UserData",
				Blocks: []models.Block{
					{
						Id:        "user_profile",
						StartCell: "A1",
						Columns: []models.Column{
							{Name: "name"},
							{Name: "email"},
						},
					},
					{
						Id:        "user_roles",
						StartCell: "D5",
						Columns: []models.Column{
							{Name: "role"},
							{Name: "level"},
						},
					},
				},
			},
		},
	}

	importer := NewImporter(bytes.NewReader(excelFileBytes.Bytes()), request)
	result, err := importer.Process()

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Empty(t, result.Errors)

	sheetData, ok := result.Data["UserData"]
	require.True(t, ok)

	userProfileData, ok := sheetData["user_profile"]
	require.True(t, ok)
	require.Len(t, userProfileData, 1)
	assert.Equal(t, "John Doe", userProfileData[0]["name"])

	userRolesData, ok := sheetData["user_roles"]
	require.True(t, ok)
	require.Len(t, userRolesData, 2)
	assert.Equal(t, "Admin", userRolesData[0]["role"])
	assert.Equal(t, "2", userRolesData[1]["level"])
}

// func TestImporter_Process_InvalidSheet(t *testing.T) {
// 	excelFileBytes := createTestExcelFileForImport(t)

// 	request := &models.Request{
// 		Sheets: []models.Sheet{
// 			{
// 				Name: "NonExistentSheet",
// 				Blocks: []models.Block{
// 					{Id: "some_block", StartCell: "A1", Columns: []models.Column{{Name: "col1"}}},
// 				},
// 			},
// 		},
// 	}

// 	importer := NewImporter(bytes.NewReader(excelFileBytes.Bytes()), request)
// 	result, err := importer.Process()

// 	require.NoError(t, err)
// 	require.NotNil(t, result)
// 	require.Len(t, result.Errors, 1)
// 	assert.Contains(t, result.Errors[0], "sheet 'NonExistentSheet' not found")
// }
