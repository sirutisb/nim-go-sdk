/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{js,ts,jsx,tsx}",
        "./nim-chat-src/src/**/*.{js,ts,jsx,tsx}"
    ],
    theme: {
        extend: {
            colors: {
                'nim-orange': '#FF6D00',
                'nim-cream': '#F1EDE7',
                'nim-blue': '#9BC1F3',
                'nim-black': '#231F18',
                'nim-brown': '#492610',
                'nim-green': '#188A31',
            },
            fontFamily: {
                display: ['ABC Marist', 'Georgia', 'serif'],
                body: ['Helvetica Monospaced Pro', 'Helvetica', 'Arial', 'sans-serif'],
            },
        },
    },
    plugins: [],
}
