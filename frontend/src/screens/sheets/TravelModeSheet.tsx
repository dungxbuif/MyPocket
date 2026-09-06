import * as React from "react";
import { Sheet } from "../../components/ui/sheet";
import { ModalHeader } from "../../components/navigation/ModalHeader";
import { GroupedCard } from "../../components/cards/GroupedCard";
import { SwitchRow } from "../../components/cards/SwitchRow";
import { FormRowItem } from "../../components/cards/FormRowItem";

export interface TravelModeSheetProps {
  isOpen: boolean;
  onClose: () => void;
  onSave?: (enabled: boolean, eventId?: string) => void;
}

export function TravelModeSheet({ isOpen, onClose, onSave }: TravelModeSheetProps) {
  const [enabled, setEnabled] = React.useState(false);
  const [selectedEvent, setSelectedEvent] = React.useState("Chuyến đi Đà Nẵng");

  return (
    <Sheet
      isOpen={isOpen}
      onClose={onClose}
      title="Chế Độ Du Lịch"
      header={
        <ModalHeader
          title="Chế Độ Du Lịch"
          dismissLabel="Huỷ"
          onDismiss={onClose}
          rightAction={{
            label: "Xong",
            isPrimary: true,
            onClick: () => {
              onSave?.(enabled, selectedEvent);
              onClose();
            },
          }}
        />
      }
    >
      <div className="flex flex-col gap-3 py-2">
        <GroupedCard>
          <SwitchRow
            title="Bật Chế Độ Du Lịch"
            description="Khi bật, các giao dịch mới tạo sẽ tự động được gán vào sự kiện du lịch mặc định."
            checked={enabled}
            onChange={setEnabled}
          />
        </GroupedCard>

        {enabled && (
          <GroupedCard title="SỰ KIỆN DU LỊCH">
            <FormRowItem
              label="Chọn sự kiện"
              value={selectedEvent}
              onClick={() => undefined}
            />
          </GroupedCard>
        )}
      </div>
    </Sheet>
  );
}
