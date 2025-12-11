export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        // Keep dark colors as fallbacks
        dark: {
          950: '#000000',
          900: '#0a0a0a',
          800: '#171717',
          700: '#262626',
        },
      },
      backdropBlur: {
        xs: '2px',
      },
    },
  },
  plugins: [],
};
