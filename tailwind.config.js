module.exports = {
  purge: [],
  darkMode: false, // or 'media' or 'class'
  theme: {
    extend: {},
    truncate: {
      lines: {
          2: '2',
          3: '3',
          5: '5',
          8: '8',
      }
    }
  },
  variants: {
    extend: {},
  },
  plugins: [
    require('tailwindcss-truncate-multiline')(),
  ],
}
