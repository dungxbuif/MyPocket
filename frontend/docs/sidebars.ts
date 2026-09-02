import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    'intro',
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