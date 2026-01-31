import NimChatWrapper from "./components/NimChatWrapper";
import ExpenseTracker from "./components/ExpenseTracker";

export default function Home() {
  return (
    <>
      <ExpenseTracker />
      <NimChatWrapper />
    </>
  );
}
