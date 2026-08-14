interface GroupSearchResultsProps {
  groupId: string;
  onOpen: (groupId: string) => void;
}

export function GroupSearchResults({ groupId, onOpen }: GroupSearchResultsProps) {
  return (
    <ul className="group-search-list">
      <li key={groupId} className="group-search-item">
        <button type="button" className="contact-search-button" onClick={() => onOpen(groupId)}>
          <span className="contact-avatar">{groupId.slice(0, 1).toUpperCase()}</span>
          <span className="contact-main">
            <span className="contact-name">{groupId}</span>
          </span>
        </button>
      </li>
    </ul>
  );
}
