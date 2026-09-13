import { BaseBarChart, BaseDonutChart } from "../molecules/BaseCharts";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { SectionTitle } from "../atoms/SectionTitle";
import { BaseCategoryTree, type BaseCategoryTreeItem } from "../molecules/BaseCategoryTree";
import { categoryShares, categoryTree, reportBars, type MockCategoryNode } from "../data/mockFinance";
import { formatVND } from "../utils/format";

const REPORT_TREE_TEXT = {
  activity: (count: number) => `Hoạt động trong ${count} nhóm`,
} as const;

export function ReportsPanel({ masked }: { masked: boolean }) {
  return (
    <>
      <SurfaceCard padding="md" className="">
        <SectionTitle title="Tổng quan báo cáo" action="Tháng 09" />
        <BaseBarChart values={reportBars} dangerFrom={8} label="Tổng quan báo cáo" />
        <div className="mt-4 grid grid-cols-3 gap-2 text-center">
          <SurfaceCard padding="sm" tone="muted" elevation="flat" className="">
            <Text size="xs" tone="secondary" className="">Chi tiêu</Text>
            <Text numeric size="sm" weight="bold" tone="danger" className="mt-1">{masked ? "••••••" : "7.830.000 đ"}</Text>
          </SurfaceCard>
          <SurfaceCard padding="sm" tone="muted" elevation="flat" className="">
            <Text size="xs" tone="secondary" className="">Thu nhập</Text>
            <Text numeric size="sm" weight="bold" tone="action" className="mt-1">{masked ? "••••••" : "28.000.000 đ"}</Text>
          </SurfaceCard>
          <SurfaceCard padding="sm" tone="muted" elevation="flat" className="">
            <Text size="xs" tone="secondary" className="">Dự báo</Text>
            <Text size="sm" weight="bold" tone="ink" className="mt-1">+2.1M</Text>
          </SurfaceCard>
        </div>
      </SurfaceCard>

      <SurfaceCard padding="md" className="">
        <SectionTitle title="Cơ cấu chi tiêu" />
        <BaseDonutChart shares={categoryShares} center="Ăn uống" />
      </SurfaceCard>

      <SurfaceCard padding="md" className="">
        <SectionTitle title="Danh mục đa tầng" />
        <div className="mt-3 space-y-3">{categoryTree.map((node) => <BaseCategoryTree key={node.id} root={reportTreeItem(node, masked, true)} children={(node.children ?? []).map((child) => reportTreeItem(child, masked, false))} />)}</div>
      </SurfaceCard>
    </>
  );
}

function reportTreeItem(node: MockCategoryNode, masked: boolean, root: boolean): BaseCategoryTreeItem {
  return {
    id: node.id,
    name: node.name,
    subtitle: root ? REPORT_TREE_TEXT.activity(node.children?.length ?? 0) : "",
    isSystem: true,
    icon: node.icon,
    trailing: <span className={`money font-bold text-ink ${root ? "text-sm" : "text-xs"}`}>{masked ? "••••••" : formatVND(node.amount)}</span>,
  };
}
