import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    'intro',
    'updates/2026-09-10-audit',
    'updates/2026-09-09',
    {
      type: 'category',
      label: 'Skills cho AI agents',
      items: ['skills/business-logic-audit'],
    },
    {
      type: 'category',
      label: 'Cơ sở dữ liệu (ERD)',
      items: [
        'erd/overview',
        'erd/entities',
        'erd/relationships',
      ],
    },
    {
      type: 'category',
      label: 'API',
      items: [
        'api/overview',
        'api/authentication',
        'api/wallets',
        'api/categories',
        'api/transactions',
        'api/budgets',
        'api/planning',
        'api/analytics',
        'api/assets',
        'api/sync',
        'api/notifications',
        'api/export',
      ],
    },
  ],
};

export default sidebars;
