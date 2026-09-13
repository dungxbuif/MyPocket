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
            <h2 className="mt-3 text-base font-bold tracking-tight text-[#1b1c1c]">
               {name}
            </h2>
            <p className="text-xs font-medium text-[#6f7a6b]">{email}</p>
         </div>
         {onManage ? (
            <button
               type="button"
               className={PROFILE_CARD_CLASSES.row}
               onClick={onManage}
            >
               <span className="flex items-center gap-3">
                  <span className="grid h-8 w-8 place-items-center rounded-full bg-[#f5f3f3] text-[#3f4a3c]">
                     <UserRound size={17} />
                  </span>
                  <span>
                     <strong className="block text-sm text-[#1b1c1c]">
                        Quản lý tài khoản
                     </strong>
                     <small className="text-[11px] text-[#6f7a6b]">
                        Hồ sơ Google
                     </small>
                  </span>
               </span>
               <span aria-hidden="true" className="text-lg text-[#6f7a6b]">
                  ›
               </span>
            </button>
         ) : null}
      </SurfaceCard>
   );
}
