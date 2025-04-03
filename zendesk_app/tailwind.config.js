const garden = require('@zendeskgarden/tailwindcss');

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './src/**/*.{html,js,vue}',
  ],
  theme: {
    extend: {},
  },
  plugins: [
    garden({
      prefix: 'g-'
    })
  ],
}
