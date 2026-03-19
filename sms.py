import requests

PHONE_URL = "http://192.168.100.165:8080/send-sms"

def send_sms(phone, message):
    payload = {
        "phone": phone,
        "message": message
    }
    try:
        r = requests.post(PHONE_URL, json=payload)
        print(f"SMS sent to {phone}: {r.status_code}")
        print(r.text)
    except Exception as e:
        print(f"Failed: {e}")


# --- Example usage ---
if __name__ == "__main__":
    name = "MOIST"
    phone = "09856122843"
    otp = "123456"
    
    # Unlimited SMS sender
    print("=== UNLIMITED SMS SENDER ===")
    try:
        count = int(input("How many messages to send? "))
        delay = float(input("Delay between messages (seconds)? "))
        
        for i in range(count):
            send_sms(phone, f"{name}: Your OTP is {otp} [{i+1}/{count}]")
            if i < count - 1:  # Don't delay after last message
                import time
                time.sleep(delay)
                
        print(f"\n✅ Successfully sent {count} messages!")
        
    except ValueError:
        print("Invalid input. Please enter numbers only.")
    except KeyboardInterrupt:
        print("\n⏹️  Sending stopped by user.")
