import asyncio
import aioglobal_hotkeys.aioglobal_hotkeys as hotkeys

is_running = True

def shutdown():
    global is_running
    is_running = False

async def main():
    hotkeys.register_hotkeys([
        [["control", "shift", "q"], None, shutdown],
        [["control", "shift", "5"], lambda: print("Pressed"), lambda: print("Released")]
    ])

    hotkeys.start_checking_hotkeys()

    while is_running:
        await asyncio.sleep(0.1)

    hotkeys.stop_checking_hotkeys()

if __name__ == "__main__":
    asyncio.run(main())
