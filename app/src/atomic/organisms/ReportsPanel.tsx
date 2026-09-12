import { SectionTitle } from "../atoms/SectionTitle";
import { CategoryTreeCard } from "../molecules/CategoryTreeCard";
import { categoryShares, categoryTree, reportBars } from "../data/mockFinance";

export function ReportsPanel({ masked }: { masked: boolean }) {
  return (
    <>
      <section className="rounded-3xl bg-white p-4">
        <SectionTitle title="Tổng quan báo cáo" action="Tháng 09" />
        <div className="mt-4 flex h-32 items-end gap-2">
          {reportBars.map((bar, index) => (
            <div key={index} className="flex flex-1 items-end">
              <div className="w-full rounded-t-full bg-[#e3e2e2]" style={{ height: `${bar}%` }}>
                <div className={`rounded-t-full ${index > 7 ? "bg-[#bb1614]" : "bg-[#006e1c]"}`} style={{ height: "55%" }} />
              </div>
            </div>
          ))}
        </div>
        <div className="mt-4 grid grid-cols-3 gap-2 text-center">
          <div className="rounded-2xl bg-[#f5f3f3] p-3">
            <p className="text-xs text-[#3f4a3c]">Chi tiêu</p>
            <p className="money mt-1 text-sm font-bold text-[#bb1614]">{masked ? "••••••" : "7.830.000 đ"}</p>
          </div>
          <div className="rounded-2xl bg-[#f5f3f3] p-3">
            <p className="text-xs text-[#3f4a3c]">Thu nhập</p>
            <p className="money mt-1 text-sm font-bold text-[#006e1c]">{masked ? "••••••" : "28.000.000 đ"}</p>
          </div>
          <div className="rounded-2xl bg-[#f5f3f3] p-3">
            <p className="text-xs text-[#3f4a3c]">Dự báo</p>
            <p className="mt-1 text-sm font-bold text-[#1b1c1c]">+2.1M</p>
          </div>
        </div>
      </section>

      <section className="rounded-3xl bg-white p-4">
        <SectionTitle title="Cơ cấu chi tiêu" />
        <div className="mt-4 flex items-center gap-5">
          <div
            className="h-32 w-32 shrink-0 rounded-full"
            style={{
              background: `conic-gradient(${categoryShares
                .reduce<Array<string>>((parts, item, index) => {
                  const start = categoryShares.slice(0, index).reduce((sum, share) => sum + share.value, 0);
                  const end = start + item.value;
                  parts.push(`${item.color} ${start}% ${end}%`);
                  return parts;
                }, [])
                .join(", ")})`,
            }}
          >
            <div className="grid h-full w-full place-items-center rounded-full p-5">
              <div className="grid h-20 w-20 place-items-center rounded-full bg-white text-center">
                <span className="text-xs font-bold text-[#3f4a3c]">Top</span>
                <strong className="text-sm">Ăn uống</strong>
              </div>
            </div>
          </div>
          <div className="min-w-0 flex-1 space-y-2">
            {categoryShares.map((share) => (
              <div key={share.name} className="flex items-center gap-2 text-sm">
                <span className="h-3 w-3 rounded-full" style={{ background: share.color }} />
                <span className="min-w-0 flex-1 truncate">{share.name}</span>
                <strong>{share.value}%</strong>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="rounded-3xl bg-white p-4">
        <SectionTitle title="Danh mục đa tầng" />
        <div className="mt-3 space-y-3">{categoryTree.map((node) => <CategoryTreeCard key={node.id} node={node} masked={masked} />)}</div>
      </section>
    </>
  );
}
