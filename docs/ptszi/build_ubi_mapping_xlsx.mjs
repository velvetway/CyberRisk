import fs from "node:fs/promises";
import path from "node:path";
import { SpreadsheetFile, Workbook } from "@oai/artifact-tool";

const dir = path.dirname(new URL(import.meta.url).pathname);
const csvPath = path.join(dir, "ubi_mapping.csv");
const outputPath = path.join(dir, "ptszi_ubi_mapping.xlsx");

function parseCSV(text) {
  const rows = [];
  let row = [];
  let cell = "";
  let quoted = false;
  for (let i = 0; i < text.length; i++) {
    const ch = text[i];
    const next = text[i + 1];
    if (quoted) {
      if (ch === '"' && next === '"') {
        cell += '"';
        i++;
      } else if (ch === '"') {
        quoted = false;
      } else {
        cell += ch;
      }
      continue;
    }
    if (ch === '"') {
      quoted = true;
    } else if (ch === ",") {
      row.push(cell);
      cell = "";
    } else if (ch === "\n") {
      row.push(cell);
      rows.push(row);
      row = [];
      cell = "";
    } else if (ch !== "\r") {
      cell += ch;
    }
  }
  if (cell.length || row.length) {
    row.push(cell);
    rows.push(row);
  }
  return rows;
}

const csv = await fs.readFile(csvPath, "utf8");
const rows = parseCSV(csv);
const headers = rows[0];
const data = rows.slice(1);

const workbook = Workbook.create();
const mapping = workbook.worksheets.add("УБИ-ST-VL");
mapping.getRangeByIndexes(0, 0, 1, headers.length).values = [headers];
mapping.getRangeByIndexes(1, 0, data.length, headers.length).values = data;

mapping.getRangeByIndexes(0, 0, 1, headers.length).format = {
  fill: "#1F4E78",
  font: { bold: true, color: "#FFFFFF" },
};

const widths = [90, 340, 260, 220, 90, 80, 260, 42, 42, 42, 80, 260, 360, 340, 300];
for (let i = 0; i < widths.length; i++) {
  mapping.getRangeByIndexes(0, i, data.length + 1, 1).format.columnWidthPx = widths[i];
}
mapping.getRangeByIndexes(0, 0, data.length + 1, headers.length).format.wrapText = true;
mapping.getRangeByIndexes(1, 5, data.length, 1).format.numberFormat = "0.00";
mapping.getRangeByIndexes(1, 10, data.length, 1).format.numberFormat = "0.00";
mapping.tables.add(`A1:O${data.length + 1}`, true, "UbiMappingTable");

const legend = workbook.worksheets.add("Правила");
legend.getRange("A1:D1").values = [["Блок", "Правило", "Зачем", "Пример"]];
legend.getRange("A2:D12").values = [
  ["УБИ", "Каждая запись БДУ ФСТЭК рассматривается как конкретная угроза ST_i", "Сохраняем 227 угроз без потери детализации", "УБИ.030"],
  ["Укрупнение ST", "УБИ дополнительно привязывается к одной или нескольким группам ST", "Группа нужна для матрицы VL/методов и компактного UI", "УБИ.030 -> ST2, ST7"],
  ["Источники S", "Внутренний нарушитель -> S3; внешний низкий -> S4; внешний средний/высокий -> S1/S2/S4", "Сопоставление формулировок БДУ с источниками исходной схемы", "Внешний средний -> S1/S2/S4"],
  ["Q threat", "низкий=0.33, средний=0.66, высокий=1.00", "Равномерная нормализация потенциала нарушителя", "medium -> 0.66"],
  ["q threat", "1 свойство К/Ц/Д=0.33, 2=0.66, 3=1.00", "Равномерная нормализация последствий", "К+Ц+Д -> 1.00"],
  ["VL", "VL берутся из матрицы группы ST -> VL", "Определяет, через какие звенья реализуется угроза", "ST7 -> VL2/VL4/VL6/VL7"],
  ["Методы", "Методы берутся из матрицы VL -> controls", "Определяет, чем закрывать уязвимые звенья", "VL7 -> FW/IDS/DZ/DD/..."],
  ["Q reaction", "Считается по actual_VL = VL(угрозы) ∩ VL(актива)", "Учитывается только то, что актуально для выбранного актива", "VL2, VL6, VL7"],
  ["W", "W = (Q + q + (1 - Qreaction)) / 3 * Z", "Итоговый риск сценария", "чем выше Qreaction, тем ниже W"],
  ["Важно", "Одна УБИ может иметь несколько ST-групп", "Одна угроза ФСТЭК часто реализует несколько сценариев", "XSS -> ST2 и ST7"],
  ["UI", "Админ видит риск, VL, методы и УБИ без служебных коэффициентов", "Коэффициенты остаются в таблице/API для обоснования", "страница /ptszi/model"],
];
legend.getRange("A1:D1").format = { fill: "#1F4E78", font: { bold: true, color: "#FFFFFF" } };
legend.getRange("A1:D12").format.wrapText = true;
for (const [idx, width] of [260, 420, 420, 260].entries()) {
  legend.getRangeByIndexes(0, idx, 12, 1).format.columnWidthPx = width;
}
legend.tables.add("A1:D12", true, "RulesTable");

const xlsx = await SpreadsheetFile.exportXlsx(workbook);
await xlsx.save(outputPath);
console.log(outputPath);
