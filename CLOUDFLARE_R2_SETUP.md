# How to Set Up Cloudflare R2 for Photobox

This guide explains how to get the 5 pieces of information required in the **Photo Sharing (Cloudflare R2)** section of the Photobox Admin Panel.

Cloudflare R2 is an object storage solution that will host the photos taken by the photobooth. When a session ends, the app uploads the image to R2 and generates a QR code for your guests to scan.

> **Note:** A Cloudflare account is free, and R2 is free for up to 10 GB of storage/month. However, you must attach a payment method to your Cloudflare account to enable R2.

---

## 1. Account ID
Your Account ID is a unique string representing your Cloudflare account.

1. Log in to your Cloudflare dashboard at [dash.cloudflare.com](https://dash.cloudflare.com).
2. Look at the URL in your web browser's address bar. 
3. The URL will look like `https://dash.cloudflare.com/2a63bbbcc87cbd5bdb...`
4. The long string between `dash.cloudflare.com/` and the next `/` is your **Account ID**.
5. Copy this string and paste it into the **ACCOUNT ID** box in the admin panel.

## 2. Bucket Name
A "bucket" is the specific storage folder where your photos will go.

1. On the left sidebar of the Cloudflare dashboard, under **Storage & databases**, click **R2 object storage**.
2. Click the blue **Create bucket** button on the right side of the screen.
3. Enter a name (e.g., `photobox`). It must be lowercase and unique.
4. Leave the default settings (Standard storage class, etc.) and click **Create bucket**.
5. The name you typed is your **Bucket Name**. Type this exactly into the admin panel.

## 3. Public Base URL
You need to make the bucket public so guests' phones can download the image when they scan the QR code.

1. Click into the bucket you just created (e.g., `photobox`).
2. Go to the **Settings** tab.
3. Look for the **Public Access** section. Under *R2.dev subdomain* (or *Public Development URL*), click **Allow Access**.
4. Type `allow` in the prompt to confirm.
5. A green URL will appear (e.g., `https://pub-044dfe1...r2.dev`).
6. Copy this entire URL into the **PUBLIC BASE URL** box. *(Make sure not to include a trailing slash `/` at the end).*

## 4. Access Key ID & Secret Access Key
These keys give the Photobox app permission to upload files into your bucket on your behalf.

1. Go back to the main **R2 Object Storage** overview page (click it on the left sidebar).
2. On the right side of the screen, look for the button that says **Manage R2 API Tokens** and click it.
3. Click the blue **Create Account API token** button.
4. **Token name:** Give it a memorable name (e.g., "Photobox Upload").
5. **Permissions:** Change this to **Admin Read & Write** (or **Object Read & Write**) so the app has permission to upload photos.
6. Scroll to the bottom and click **Create API Token**.
7. Keep this page open! You will see two long strings:
   - **Access Key ID:** Copy this and paste it into the **ACCESS KEY ID** box.
   - **Secret Access Key:** Copy this and paste it into the **SECRET ACCESS KEY** box. 

> **WARNING:** Cloudflare will only show you the Secret Access Key **once**. Do not close the page until you have saved it in the admin panel. If you lose it, you will have to create a new token.

---

Once you have filled out all 5 boxes in the Photobox Admin Panel, click **Save R2 Settings**. Your photobooth is now ready to upload photos and generate QR codes!
