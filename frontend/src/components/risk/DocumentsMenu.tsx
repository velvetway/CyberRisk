// DocumentsMenu — выпадающее меню «Скачать документы» в карточке актива.
// Закрывает блок «Формирование организационно-технической документации»
// из 7.png диплома: одной кнопкой генерирует пакет PDF
// (паспорт АС, модель угроз, план защиты, опционально — отчёт ОСЗ).
import React, { useEffect, useRef, useState } from "react";
import { authFetch } from "../../api/client";
import { Btn, Icon } from "../design";

interface Props {
  assetID: number;
}

interface DocOption {
  key: string;
  label: string;
  description: string;
  url: (id: number) => string;
  filename: (id: number) => string;
  highlight?: boolean;
}

const DOC_OPTIONS: DocOption[] = [
  {
    key: "pack",
    label: "Полный пакет (ZIP)",
    description: "Все 4 документа в одном архиве",
    url: (id) => `/api/reports/asset/${id}/documents.zip`,
    filename: (id) => `documents-asset-${id}.zip`,
    highlight: true,
  },
  {
    key: "passport",
    label: "Технический паспорт АС",
    description: "Общие сведения, состав ПО, СЗИ, соответствие",
    url: (id) => `/api/reports/asset/${id}/document/passport.pdf`,
    filename: (id) => `passport-asset-${id}.pdf`,
  },
  {
    key: "threat-model",
    label: "Модель угроз ИБ",
    description: "Применимые угрозы, цепочки атак S→ST→VL→DA, расчёт W",
    url: (id) => `/api/reports/asset/${id}/document/threat-model.pdf`,
    filename: (id) => `threat-model-asset-${id}.pdf`,
  },
  {
    key: "protection-plan",
    label: "Перечень мер защиты",
    description: "Контроли по 4 группам АРМ/ЛВС/ЭДО/конф.инф. + рекомендации",
    url: (id) => `/api/reports/asset/${id}/document/protection-plan.pdf`,
    filename: (id) => `protection-plan-asset-${id}.pdf`,
  },
  {
    key: "compliance",
    label: "Отчёт о состоянии защищённости (ОСЗ)",
    description: "% соответствия ФСТЭК-17 и ISO 27001 по требованиям",
    url: (id) => `/api/compliance/asset/${id}/report.pdf`,
    filename: (id) => `compliance-asset-${id}.pdf`,
  },
];

export const DocumentsMenu: React.FC<Props> = ({ assetID }) => {
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState<string | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const wrapRef = useRef<HTMLDivElement>(null);

  // Закрытие меню по клику вне.
  useEffect(() => {
    if (!open) return;
    const onClick = (e: MouseEvent) => {
      if (wrapRef.current && !wrapRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", onClick);
    return () => document.removeEventListener("mousedown", onClick);
  }, [open]);

  const download = async (opt: DocOption) => {
    setErr(null);
    setBusy(opt.key);
    try {
      const res = await authFetch(opt.url(assetID));
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = opt.filename(assetID);
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
      setOpen(false);
    } catch (e) {
      setErr((e as Error).message);
    } finally {
      setBusy(null);
    }
  };

  return (
    <div ref={wrapRef} style={{ position: "relative" }}>
      <Btn variant="outline" onClick={() => setOpen((v) => !v)} icon={<Icon name="file" />}>
        Документы {open ? "▴" : "▾"}
      </Btn>

      {open && (
        <div
          style={{
            position: "absolute",
            top: "calc(100% + 6px)",
            right: 0,
            background: "var(--bg-elev-1)",
            border: "1px solid var(--border)",
            borderRadius: "var(--r-md)",
            boxShadow: "var(--shadow-lg, 0 8px 24px rgba(0,0,0,0.18))",
            minWidth: 360,
            padding: 6,
            zIndex: 50,
            display: "flex",
            flexDirection: "column",
            gap: 2,
          }}
        >
          <div style={{ fontSize: 11, color: "var(--fg-faint)", padding: "6px 10px 4px", textTransform: "uppercase", letterSpacing: 0.4 }}>
            Орг.-тех. документация
          </div>
          {DOC_OPTIONS.map((opt) => (
            <button
              key={opt.key}
              onClick={() => download(opt)}
              disabled={busy !== null}
              style={{
                background: opt.highlight ? "var(--accent-ghost)" : "transparent",
                color: "var(--fg)",
                border: "none",
                textAlign: "left",
                padding: "8px 10px",
                borderRadius: "var(--r-sm)",
                cursor: busy === null ? "pointer" : "wait",
                opacity: busy && busy !== opt.key ? 0.5 : 1,
                display: "flex",
                flexDirection: "column",
                gap: 2,
              }}
              onMouseEnter={(e) => {
                if (!busy) (e.currentTarget as HTMLButtonElement).style.background = "var(--bg-elev-2)";
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLButtonElement).style.background = opt.highlight ? "var(--accent-ghost)" : "transparent";
              }}
            >
              <span style={{ fontSize: 13, fontWeight: 600 }}>
                {opt.label}
                {busy === opt.key && <span style={{ marginLeft: 6, color: "var(--fg-faint)" }}>…</span>}
              </span>
              <span style={{ fontSize: 11, color: "var(--fg-faint)" }}>{opt.description}</span>
            </button>
          ))}
          {err && (
            <div style={{ color: "var(--risk-critical)", fontSize: 12, padding: "4px 10px" }}>{err}</div>
          )}
        </div>
      )}
    </div>
  );
};
