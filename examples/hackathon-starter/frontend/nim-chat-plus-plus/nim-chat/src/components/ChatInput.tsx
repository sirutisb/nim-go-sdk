import { useState, useRef, useEffect, FormEvent, KeyboardEvent } from 'react';

// Hardcoded list of users available for tagging
const TAGGABLE_USERS = [
  { id: '0', username: 'jack', displayName: 'Jack Marsh' },
  { id: '1', username: 'money', displayName: 'Benitas Sirutis' },
  { id: '2', username: 'quant', displayName: 'Thomas Nguyen' },
  { id: '3', username: 'kenna2266', displayName: 'Kenna' },
  { id: '4', username: 'kk598', displayName: 'Kazybek' },
  { id: '5', username: 'alpha', displayName: 'Marcus Thorne' },
  { id: '6', username: 'cipher99', displayName: 'Elena' },
  { id: '7', username: 'vortex', displayName: 'Julian Rossi' },
  { id: '8', username: 'amara4412', displayName: 'Amara' },
  { id: '9', username: 'glitch', displayName: 'Soren Bjorn' },
  { id: '10', username: 'lyra8821', displayName: 'Lyra' },
  { id: '11', username: 'macro', displayName: 'David Chen' },
  { id: '12', username: 'st8829', displayName: 'Stefan' },
  { id: '13', username: 'nexus', displayName: 'Isabella Varga' },
  { id: '14', username: 'jj1029', displayName: 'Jesper' },
  { id: '15', username: 'pixel', displayName: 'Aria Montgomery' },
  { id: '16', username: 'omega', displayName: 'Klaus Schmidt' },
  { id: '17', username: 'mira7733', displayName: 'Mira' },
  { id: '18', username: 'tango', displayName: 'Victor Vance' },
  { id: '19', username: 'rk5521', displayName: 'Rakesh' },
  { id: '20', username: 'solver', displayName: 'Naomi Wattson' },
  { id: '21', username: 'zenith', displayName: 'Hiroshi Tanaka' },
  { id: '22', username: 'lo2291', displayName: 'Lola' },
  { id: '23', username: 'hazard', displayName: 'Dante Alighieri' },
  { id: '24', username: 'bb9910', displayName: 'Bibi' },
  { id: '25', username: 'logic', displayName: 'Sarah Jenkins' },
  { id: '26', username: 'flux', displayName: 'Arjun Mehta' },
  { id: '27', username: 'nyx4490', displayName: 'Nyx' },
  { id: '28', username: 'proxy', displayName: 'Oliver Quinn' },
  { id: '29', username: 'mm8271', displayName: 'Marek' },
  { id: '30', username: 'atlas', displayName: 'Cassandra Cain' },
  { id: '31', username: 'vector', displayName: 'Leo Maxwell' },
  { id: '32', username: 'whop', displayName: 'Iman Gadzhi' },
  { id: '33', username: 'orbit', displayName: 'Felix Wright' },
  { id: '34', username: 'pp5562', displayName: 'Pavel' },
  { id: '35', username: 'titan', displayName: 'Beatrice Webb' },
  { id: '36', username: 'neon', displayName: 'Zoe Kravitz' },
  { id: '37', username: 'kai8827', displayName: 'Kai' },
  { id: '38', username: 'delta', displayName: 'Simon Pegg' },
  { id: '39', username: 'gg4491', displayName: 'Gigi' },
  { id: '40', username: 'prime', displayName: 'Arthur Pendragon' },
  { id: '41', username: 'shadow', displayName: 'Malik Williams' },
  { id: '42', username: 'li7721', displayName: 'Liana' },
  { id: '43', username: 'echo', displayName: 'Sophie Turner' },
  { id: '44', username: 'tt9012', displayName: 'Tariq' },
  { id: '45', username: 'apex', displayName: 'Chloe Zhao' },
  { id: '46', username: 'sonic', displayName: 'Miles Prower' },
  { id: '47', username: 'erika332', displayName: 'Erika' },
  { id: '48', username: 'ghost', displayName: 'Liam Neeson' },
  { id: '49', username: 'kk1128', displayName: 'Kirill' },
  { id: '50', username: 'signal', displayName: 'Fiona Gallagher' },
  { id: '51', username: 'pulse', displayName: 'Oscar Isaac' },
  { id: '52', username: 'mara229', displayName: 'Mara' },
  { id: '53', username: 'static', displayName: 'Virgil Hawkins' },
  { id: '54', username: 'vv8831', displayName: 'Vadim' },
  { id: '55', username: 'hyper', displayName: 'Penny Parker' },
  { id: '56', username: 'prism', displayName: 'Silas Vane' },
  { id: '57', username: 'hugo112', displayName: 'Hugo' },
  { id: '58', username: 'binary', displayName: 'Ada Lovelace' },
  { id: '59', username: 'ss7742', displayName: 'Sasha' },
  { id: '60', username: 'optic', displayName: 'Scott Summers' },
  { id: '61', username: 'void', displayName: 'Evelyn Wang' },
  { id: '62', username: 'lara441', displayName: 'Lara' },
  { id: '63', username: 'cipher', displayName: 'Ramsey Bolton' },
  { id: '64', username: 'study', displayName: 'Luke Belmar' },
  { id: '66', username: 'rally', displayName: 'Ken Block' },
  { id: '67', username: 'ivan009', displayName: 'Ivan' },
  { id: '68', username: 'summit', displayName: 'Hillary Tenzing' },
  { id: '69', username: 'aa2218', displayName: 'Anish' },
  { id: '70', username: 'spark', displayName: 'Barry Allen' },
  { id: '71', username: 'motion', displayName: 'Eadweard Muybridge' },
  { id: '72', username: 'tina661', displayName: 'Tina' },
  { id: '73', username: 'radar', displayName: 'Alan Alda' },
  { id: '74', username: 'rr5561', displayName: 'Rory' },
  { id: '75', username: 'focus', displayName: 'Will Smith' },
  { id: '76', username: 'metric', displayName: 'Isaac Newton' },
  { id: '77', username: 'yara992', displayName: 'Yara' },
  { id: '78', username: 'vault', displayName: 'Robert Hanssen' },
  { id: '79', username: 'dd1109', displayName: 'Dmitry' },
  { id: '80', username: 'shift', displayName: 'Lewis Hamilton' },
  { id: '81', username: 'buffer', displayName: 'Linus Torvalds' },
  { id: '82', username: 'kira881', displayName: 'Kira' },
  { id: '83', username: 'syntax', displayName: 'Guido van Rossum' },
  { id: '84', username: 'ff4431', displayName: 'Finn' },
  { id: '85', username: 'module', displayName: 'Margaret Hamilton' },
  { id: '86', username: 'array', displayName: 'Donald Knuth' },
  { id: '87', username: 'otto221', displayName: 'Otto' },
  { id: '88', username: 'scalar', displayName: 'Blaise Pascal' },
  { id: '89', username: 'jj7712', displayName: 'Jana' },
  { id: '90', username: 'quartz', displayName: 'Crystal Waters' },
  { id: '91', username: 'plasma', displayName: 'Nikola Tesla' },
  { id: '92', username: 'olga002', displayName: 'Olga' },
  { id: '93', username: 'fusion', displayName: 'Marie Curie' },
  { id: '94', username: 'bb3391', displayName: 'Bastian' },
  { id: '95', username: 'carbon', displayName: 'Dorothy Hodgkin' },
  { id: '96', username: 'silicon', displayName: 'Gordon Moore' },
  { id: '97', username: 'tess119', displayName: 'Tess' },
  { id: '98', username: 'cobalt', displayName: 'Grace Hopper' },
  { id: '99', username: 'll2281', displayName: 'Luka' },
  { id: '100', username: 'copper', displayName: 'James Maxwell' },
  { id: '101', username: 'iron', displayName: 'Tony Stark' },
  { id: '102', username: 'macy441', displayName: 'Macy' },
  { id: '103', username: 'steel', displayName: 'John Henry' },
  { id: '104', username: 'nn9912', displayName: 'Nico' }
];

// Anonymous avatar component
function AnonAvatar({ size = 32 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 32 32"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="rounded-full bg-nim-cream"
    >
      <circle cx="16" cy="16" r="16" fill="#F5F0EB" />
      <circle cx="16" cy="12" r="5" fill="#C4B5A6" />
      <path
        d="M6 28C6 22.4772 10.4772 18 16 18C21.5228 18 26 22.4772 26 28"
        stroke="#C4B5A6"
        strokeWidth="2"
        fill="none"
      />
    </svg>
  );
}

// Transaction type for attached transaction
interface AttachedTransaction {
  id: string;
  amount: string;
  currency: string;
  direction: string;
  note: string;
  type: string;
  counterparty: string;
  created_at: string;
}

interface ChatInputProps {
  onSend: (message: string) => void;
  disabled?: boolean;
  placeholder?: string;
  attachedTransaction?: AttachedTransaction | null;
  onRemoveAttachment?: () => void;
}

export function ChatInput({
  onSend,
  disabled = false,
  placeholder = 'Type a message...',
  attachedTransaction,
  onRemoveAttachment,
}: ChatInputProps) {
  const [value, setValue] = useState('');
  const [showMentions, setShowMentions] = useState(false);
  const [mentionQuery, setMentionQuery] = useState('');
  const [mentionStartIndex, setMentionStartIndex] = useState(-1);
  const [selectedMentionIndex, setSelectedMentionIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const mentionListRef = useRef<HTMLDivElement>(null);

  // Filter users based on mention query
  const filteredUsers = TAGGABLE_USERS.filter(
    (user) =>
      user.username.toLowerCase().includes(mentionQuery.toLowerCase()) ||
      user.displayName.toLowerCase().includes(mentionQuery.toLowerCase())
  );

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (value.trim() && !disabled) {
      onSend(value);
      setValue('');
      setShowMentions(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    const cursorPos = e.target.selectionStart || 0;
    setValue(newValue);

    // Check for @ mention trigger
    const textBeforeCursor = newValue.slice(0, cursorPos);
    const lastAtIndex = textBeforeCursor.lastIndexOf('@');

    if (lastAtIndex !== -1) {
      const textAfterAt = textBeforeCursor.slice(lastAtIndex + 1);
      // Check if there's a space after @, which means the mention is complete
      if (!textAfterAt.includes(' ')) {
        setMentionQuery(textAfterAt);
        setMentionStartIndex(lastAtIndex);
        setShowMentions(true);
        setSelectedMentionIndex(0);
        return;
      }
    }

    setShowMentions(false);
    setMentionQuery('');
    setMentionStartIndex(-1);
  };

  const handleSelectUser = (user: typeof TAGGABLE_USERS[0]) => {
    if (mentionStartIndex === -1) return;

    // Replace the @query with @username
    const beforeMention = value.slice(0, mentionStartIndex);
    const afterMention = value.slice(mentionStartIndex + mentionQuery.length + 1);
    const newValue = `${beforeMention}@${user.username} ${afterMention}`;

    setValue(newValue);
    setShowMentions(false);
    setMentionQuery('');
    setMentionStartIndex(-1);

    // Focus back on input
    inputRef.current?.focus();
  };

  const handleKeyDown = (e: KeyboardEvent) => {
    if (showMentions && filteredUsers.length > 0) {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setSelectedMentionIndex((prev) =>
          prev < filteredUsers.length - 1 ? prev + 1 : 0
        );
        return;
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault();
        setSelectedMentionIndex((prev) =>
          prev > 0 ? prev - 1 : filteredUsers.length - 1
        );
        return;
      }
      if (e.key === 'Tab' || e.key === 'Enter') {
        e.preventDefault();
        handleSelectUser(filteredUsers[selectedMentionIndex]);
        return;
      }
      if (e.key === 'Escape') {
        e.preventDefault();
        setShowMentions(false);
        return;
      }
    }

    if (e.key === 'Enter' && !e.shiftKey && !showMentions) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  // Scroll selected mention into view
  useEffect(() => {
    if (mentionListRef.current && showMentions) {
      const selectedElement = mentionListRef.current.children[selectedMentionIndex] as HTMLElement;
      if (selectedElement) {
        selectedElement.scrollIntoView({ block: 'nearest' });
      }
    }
  }, [selectedMentionIndex, showMentions]);

  // Focus input on mount
  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  // Close mentions dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        mentionListRef.current &&
        !mentionListRef.current.contains(e.target as Node) &&
        inputRef.current &&
        !inputRef.current.contains(e.target as Node)
      ) {
        setShowMentions(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <form onSubmit={handleSubmit} className="relative flex flex-col gap-2">
      {/* Attached Transaction Tag */}
      {attachedTransaction && (
        <div className="attached-transaction-tag flex items-center gap-2 px-3 py-2 bg-nim-cream rounded-lg border-2 border-nim-orange/30">
          <div className="flex-1 flex items-center gap-2">
            <div className={`w-6 h-6 rounded-full flex items-center justify-center ${attachedTransaction.direction === 'credit' ? 'bg-green-100 text-green-600' : 'bg-red-100 text-red-600'
              }`}>
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                {attachedTransaction.direction === 'credit'
                  ? <path d="M12 19V5M5 12l7-7 7 7" />
                  : <path d="M12 5v14M5 12l7 7 7-7" />
                }
              </svg>
            </div>
            <div className="flex flex-col">
              <span className="font-body text-sm font-medium text-nim-black">
                {attachedTransaction.note || `${attachedTransaction.type} transaction`}
              </span>
              <span className="font-body text-xs text-nim-brown/60">
                {attachedTransaction.direction === 'credit' ? '+' : ''}{attachedTransaction.amount} {attachedTransaction.currency} · Transaction
              </span>
            </div>
          </div>
          <button
            type="button"
            onClick={onRemoveAttachment}
            className="w-6 h-6 flex items-center justify-center rounded-full hover:bg-nim-orange/10 text-nim-brown/60 hover:text-nim-orange transition-colors"
            aria-label="Remove attachment"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
      )}

      <div className="relative flex gap-2">
        {/* Mentions dropdown */}
        {showMentions && filteredUsers.length > 0 && (
          <div
            ref={mentionListRef}
            className="absolute bottom-full left-0 right-12 mb-2 max-h-48 overflow-y-auto bg-white border-2 border-nim-cream rounded-lg shadow-lg z-50"
          >
            {filteredUsers.map((user, index) => (
              <button
                key={user.id}
                type="button"
                onClick={() => handleSelectUser(user)}
                className={`
                w-full flex items-center gap-3 px-3 py-2.5
                text-left transition-colors
                ${index === selectedMentionIndex
                    ? 'bg-nim-orange/10 text-nim-orange'
                    : 'hover:bg-nim-cream text-nim-black'
                  }
              `}
              >
                <AnonAvatar size={28} />
                <div className="flex flex-col">
                  <span className="font-body text-sm font-medium">{user.displayName}</span>
                  <span className="font-body text-xs text-nim-brown/60">@{user.username}</span>
                </div>
              </button>
            ))}
          </div>
        )}

        {/* No matches message */}
        {showMentions && mentionQuery && filteredUsers.length === 0 && (
          <div className="absolute bottom-full left-0 right-12 mb-2 px-4 py-3 bg-white border-2 border-nim-cream rounded-lg shadow-lg z-50">
            <span className="font-body text-sm text-nim-brown/60">No users found matching "@{mentionQuery}"</span>
          </div>
        )}

        <input
          ref={inputRef}
          type="text"
          value={value}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          disabled={disabled}
          placeholder={placeholder}
          className={`
          flex-1 h-11 px-4
          bg-white border-2 border-nim-cream
          rounded-lg
          text-nim-black placeholder-nim-brown/40
          font-body text-sm
          outline-none
          transition-colors duration-200
          focus:border-nim-orange
          disabled:opacity-50 disabled:cursor-not-allowed
        `}
        />
        <button
          type="submit"
          disabled={disabled || !value.trim()}
          className={`
          h-11 w-11
          bg-nim-orange text-white
          rounded-lg
          flex items-center justify-center
          transition-all duration-200
          hover:opacity-90 hover:scale-105
          active:scale-95
          disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:scale-100
        `}
          aria-label="Send message"
        >
          <svg
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <line x1="22" y1="2" x2="11" y2="13" />
            <polygon points="22 2 15 22 11 13 2 9 22 2" />
          </svg>
        </button>
      </div>
    </form>
  );
}
