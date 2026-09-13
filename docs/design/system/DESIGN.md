---
name: Financial Clarity System
colors:
  surface: '#fbf9f9'
  surface-dim: '#dbdad9'
  surface-bright: '#fbf9f9'
  surface-container-lowest: '#ffffff'
  surface-container-low: '#f5f3f3'
  surface-container: '#efeded'
  surface-container-high: '#e9e8e7'
  surface-container-highest: '#e3e2e2'
  on-surface: '#1b1c1c'
  on-surface-variant: '#3f4a3c'
  inverse-surface: '#303031'
  inverse-on-surface: '#f2f0f0'
  outline: '#6f7a6b'
  outline-variant: '#becab9'
  surface-tint: '#006e1c'
  primary: '#006e1c'
  on-primary: '#ffffff'
  primary-container: '#4caf50'
  on-primary-container: '#003c0b'
  inverse-primary: '#78dc77'
  secondary: '#556158'
  on-secondary: '#ffffff'
  secondary-container: '#d9e6da'
  on-secondary-container: '#5b675e'
  tertiary: '#bb1614'
  on-tertiary: '#ffffff'
  tertiary-container: '#ff6c5c'
  on-tertiary-container: '#6d0003'
  error: '#ba1a1a'
  on-error: '#ffffff'
  error-container: '#ffdad6'
  on-error-container: '#93000a'
  primary-fixed: '#94f990'
  primary-fixed-dim: '#78dc77'
  on-primary-fixed: '#002204'
  on-primary-fixed-variant: '#005313'
  secondary-fixed: '#d9e6da'
  secondary-fixed-dim: '#bdcabe'
  on-secondary-fixed: '#131e17'
  on-secondary-fixed-variant: '#3e4a41'
  tertiary-fixed: '#ffdad5'
  tertiary-fixed-dim: '#ffb4a9'
  on-tertiary-fixed: '#410001'
  on-tertiary-fixed-variant: '#930005'
  background: '#fbf9f9'
  on-background: '#1b1c1c'
  surface-variant: '#e3e2e2'
typography:
  display-lg:
    fontFamily: Manrope
    fontSize: 32px
    fontWeight: '700'
    lineHeight: 40px
    letterSpacing: -0.02em
  headline-md:
    fontFamily: Manrope
    fontSize: 20px
    fontWeight: '600'
    lineHeight: 28px
  headline-sm:
    fontFamily: Manrope
    fontSize: 17px
    fontWeight: '600'
    lineHeight: 24px
  body-md:
    fontFamily: Manrope
    fontSize: 15px
    fontWeight: '400'
    lineHeight: 22px
  body-sm:
    fontFamily: Manrope
    fontSize: 13px
    fontWeight: '400'
    lineHeight: 18px
  label-lg:
    fontFamily: Manrope
    fontSize: 14px
    fontWeight: '600'
    lineHeight: 20px
  label-md:
    fontFamily: Manrope
    fontSize: 12px
    fontWeight: '500'
    lineHeight: 16px
  currency-display:
    fontFamily: Manrope
    fontSize: 24px
    fontWeight: '600'
    lineHeight: 32px
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  container-margin: 1rem
  card-padding: 1.25rem
  gutter: 0.75rem
  stack-sm: 0.25rem
  stack-md: 0.5rem
  stack-lg: 1rem
---

## Brand & Style

This design system is built for personal finance management, focusing on trust, clarity, and approachability. The aesthetic leans into **Modern Corporate** with a friendly, lifestyle-oriented twist. It utilizes a clean white-label approach where data is the hero, supported by a refreshing green primary palette that symbolizes growth and stability.

The visual language is characterized by "soft modularity"—using cards with generous border radii and subtle depth to organize complex financial information into digestible "bento-box" snippets. The interface should feel organized but never rigid, encouraging users to interact with their data through tactile-feeling elements and vibrant, illustrative icons.

**Emotional Response:** Organized, secure, optimistic, and effortless.

## Colors

The palette is centered around a "Growth Green" that acts as the primary driver for actions and positive financial indicators.

*   **Primary (#4CAF50):** Used for primary buttons, progress bars representing budget health, and positive trends.
*   **Secondary (#E8F5E9):** A soft wash of the primary color, used for surface backgrounds, secondary buttons, and chip containers.
*   **Tertiary (#F44336):** Specifically reserved for negative trends, over-budget warnings, and critical alerts.
*   **Neutral (#757575):** A systematic range of grays. Use darker shades (#212121) for primary text and lighter shades (#F5F5F5) for page backgrounds and dividers.
*   **Success/Caution:** Use the Primary green for success. For neutral data (like "spent"), use a deep Slate or Navy to distinguish from functional green.

## Typography

The system uses **Manrope** for its modern, geometric balance and excellent legibility in data-dense environments. 

*   **Emphasis:** Financial figures (Currency Display) should always have a higher visual weight than their accompanying labels.
*   **Hierarchy:** Use font weight rather than just size to create distinction between primary data and metadata.
*   **Micro-copy:** Labels for "Last Month" or "Average" should use `label-md` with a neutral gray color to remain secondary to the actual values.

## Layout & Spacing

The design follows a **Fluid Grid** model with a focus on vertical stacking of cards.

*   **Margins:** A standard 16px (1rem) margin is applied to the left and right of the main screen container.
*   **Card Rhythm:** Elements within a card use an 8px base grid. Titles are separated from content by 4px (`stack-sm`), while sections within a card are separated by 12px-16px.
*   **Grouping:** Use white space rather than lines to separate logic groups where possible. When dividers are necessary, they should be 1px wide and use a very light neutral gray (#EEEEEE).
*   **Safe Areas:** Ensure the bottom navigation bar and floating action buttons (FAB) do not obscure critical transaction data.

## Elevation & Depth

Depth is used to signify interactivity and separate the background from the content layers.

*   **Surface Level (L0):** The app background is a light, neutral gray (#F8F9FA).
*   **Card Level (L1):** White surfaces (#FFFFFF) with a very soft, diffused shadow (Blur: 20px, Y: 4, Opacity: 0.05, Color: Black).
*   **Interactive Level (L2):** Elements like "Add Transaction" buttons or active bottom sheets use a slightly more pronounced shadow or a tonal tint to appear closer to the user.
*   **Backdrop Blur:** Use a subtle backdrop blur (10px-15px) for modal overlays and sticky headers to maintain context of the underlying data.

## Shapes

The shape language is consistently rounded to evoke a "friendly tech" feel.

*   **Cards:** Use `rounded-xl` (24px) for main content containers to create a soft, approachable frame.
*   **Buttons:** Standard buttons are pill-shaped (fully rounded) to maximize hit-area perception and distinguish them from card containers.
*   **Inputs:** Form fields and smaller nested elements use `rounded-lg` (12px).
*   **Progress Bars:** Always use fully rounded end-caps for a smooth, organic feel.

## Components

### Buttons
*   **Primary:** Solid Green (#4CAF50) with white text. Pill-shaped.
*   **Secondary:** Soft Green background (#E8F5E9) with Primary Green text.
*   **Ghost:** Transparent background with Primary Green text or Neutral gray for "Cancel" actions.

### Cards & Progress
*   **Budget Cards:** Include a leading icon in a circular container, a title, a primary value, and a linear progress bar. The bar should change color to Tertiary Red if the value exceeds 100%.
*   **Transaction Items:** Layout follows a `[Icon] [Title/Category] [Amount]` pattern. Amount text color should reflect transaction type (Positive = Green, Negative = Black/Neutral).

### Form Inputs
*   **Pattern:** Inset fields within a card. Labels are small and floated above the value. Use a 1px border or subtle underline to define the input area.
*   **Selection:** "Choose Category" or "Select Wallet" inputs include a trailing chevron-right icon.

### Charts
*   **Bar Charts:** Use rounded tops for bars. Current period uses Primary Green; previous periods use a muted or desaturated version.
*   **Donut/Arc Charts:** Use a medium stroke width (approx 8-10px) with rounded caps for the progress indicator.

### Icons
*   **Style:** Use illustrative, multi-color icons for categories (food, shopping) to add personality. Use simple, stroke-based icons (2px weight) for functional navigation (Home, Wallet, Reports).