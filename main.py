from bot import bot
import dotenv,os
import hellgate_watcher
import asyncio

dotenv.load_dotenv()
DISCORDTOKEN = os.getenv('DISCORDTOKEN')

def main():
    bot.run(DISCORDTOKEN)

if __name__ == "__main__":
    main()
