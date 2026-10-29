import { Text } from "../atoms/Text";
import { BaseButton } from "../atoms/BaseButton";
import { UserRound } from 'lucide-react';
import type { UserProfile } from '../../services/auth';
import { PROFILE_CARD_CLASSES } from '../atoms/tokens';
import { SurfaceCard } from '../atoms/SurfaceCard';

export function ProfileHeroCard({
   user,
   onManage,
}: {
   user: UserProfile | null;
   onManage?: () => void;
}) {
   const name = user?.name?.trim() || 'Người dùng';
   const email = user?.email?.trim() || 'Chưa có email';
   return (
      <SurfaceCard className={PROFILE_CARD_CLASSES.card}>
         <div className="flex flex-col items-center">
            <div className={PROFILE_CARD_CLASSES.avatar}>
               {name.slice(0, 1).toUpperCase()}
            </div>
            <Text as="h2" size="base" weight="bold" tone="ink" className="mt-3 tracking-tight">
               {name}
            </Text>
            <Text size="xs" weight="medium" tone="secondary" className="">{email}</Text>
         </div>
         {onManage ? (
            <BaseButton variant="row" size="row"
               type="button"
               className={PROFILE_CARD_CLASSES.row}
               onClick={onManage}
            >
               <span className="flex items-center gap-3">
                  <span className="grid h-8 w-8 place-items-center rounded-full bg-row text-secondary">
                     <UserRound size={17} />
                  </span>
                  <span>
                     <Text as="strong" size="sm" tone="ink" className="block">
                        Quản lý tài khoản
                     </Text>
                     <Text as="small" size="tiny" tone="secondary" className="">
                        Hồ sơ Google
                     </Text>
                  </span>
               </span>
               <span aria-hidden="true" className="text-lg text-secondary">
                  ›
               </span>
            </BaseButton>
         ) : null}
      </SurfaceCard>
   );
}
