export interface HighlightedToken {
  text: string;
  type: 'keyword' | 'string' | 'number' | 'comment' | 'normal';
}

const SQL_KEYWORDS = [
  'SELECT', 'FROM', 'WHERE', 'INSERT', 'UPDATE', 'DELETE', 'CREATE', 'DROP',
  'ALTER', 'TABLE', 'INDEX', 'DATABASE', 'SCHEMA', 'VIEW', 'TRIGGER',
  'PROCEDURE', 'FUNCTION', 'JOIN', 'INNER', 'LEFT', 'RIGHT', 'OUTER',
  'ON', 'AS', 'AND', 'OR', 'NOT', 'IN', 'EXISTS', 'LIKE', 'BETWEEN',
  'ORDER', 'BY', 'GROUP', 'HAVING', 'LIMIT', 'OFFSET', 'UNION', 'ALL',
  'DISTINCT', 'COUNT', 'SUM', 'AVG', 'MAX', 'MIN', 'CASE', 'WHEN', 'THEN',
  'ELSE', 'END', 'IF', 'NULL', 'IS', 'SET', 'VALUES', 'INTO', 'DEFAULT',
  'PRIMARY', 'KEY', 'FOREIGN', 'REFERENCES', 'CONSTRAINT', 'UNIQUE',
  'CHECK', 'AUTO_INCREMENT', 'AUTO_INC', 'USE', 'INDEX'
];

export function highlightSQL(sql: string): HighlightedToken[] {
  const tokens: HighlightedToken[] = [];
  let currentIndex = 0;

  // Regular expressions for different token types
  const patterns = [
    { regex: /--.*$/gm, type: 'comment' as const }, // Single line comments
    { regex: /\/\*[\s\S]*?\*\//g, type: 'comment' as const }, // Multi-line comments
    { regex: /(['"`]).*?\1/g, type: 'string' as const }, // Strings
    { regex: /\b\d+\.?\d*\b/g, type: 'number' as const }, // Numbers
  ];

  // Find all matches with their positions
  const matches: Array<{ start: number; end: number; type: HighlightedToken['type']; text: string }> = [];

  // Find comments
  const commentRegex = /(--.*$|\/\*[\s\S]*?\*\/)/gm;
  let match;
  while ((match = commentRegex.exec(sql)) !== null) {
    matches.push({
      start: match.index,
      end: match.index + match[0].length,
      type: 'comment',
      text: match[0],
    });
  }

  // Find strings
  const stringRegex = /(['"`])(?:(?=(\\?))\2.)*?\1/g;
  while ((match = stringRegex.exec(sql)) !== null) {
    matches.push({
      start: match.index,
      end: match.index + match[0].length,
      type: 'string',
      text: match[0],
    });
  }

  // Find numbers
  const numberRegex = /\b\d+\.?\d*\b/g;
  while ((match = numberRegex.exec(sql)) !== null) {
    // Check if it's not inside a string or comment
    const isInsideString = matches.some(
      m => m.type === 'string' && match!.index >= m.start && match!.index < m.end
    );
    const isInsideComment = matches.some(
      m => m.type === 'comment' && match!.index >= m.start && match!.index < m.end
    );
    if (!isInsideString && !isInsideComment) {
      matches.push({
        start: match.index,
        end: match.index + match[0].length,
        type: 'number',
        text: match[0],
      });
    }
  }

  // Find keywords
  const keywordRegex = new RegExp(`\\b(${SQL_KEYWORDS.join('|')})\\b`, 'gi');
  while ((match = keywordRegex.exec(sql)) !== null) {
    // Check if it's not inside a string or comment
    const isInsideString = matches.some(
      m => m.type === 'string' && match!.index >= m.start && match!.index < m.end
    );
    const isInsideComment = matches.some(
      m => m.type === 'comment' && match!.index >= m.start && match!.index < m.end
    );
    if (!isInsideString && !isInsideComment) {
      matches.push({
        start: match.index,
        end: match.index + match[0].length,
        type: 'keyword',
        text: match[0],
      });
    }
  }

  // Sort matches by start position
  matches.sort((a, b) => a.start - b.start);

  // Build tokens
  let lastIndex = 0;
  for (const match of matches) {
    // Add normal text before match
    if (match.start > lastIndex) {
      tokens.push({
        text: sql.substring(lastIndex, match.start),
        type: 'normal',
      });
    }
    // Add highlighted token
    tokens.push({
      text: match.text,
      type: match.type,
    });
    lastIndex = match.end;
  }

  // Add remaining text
  if (lastIndex < sql.length) {
    tokens.push({
      text: sql.substring(lastIndex),
      type: 'normal',
    });
  }

  return tokens;
}

export function renderHighlightedSQL(sql: string): string {
  const tokens = highlightSQL(sql);
  return tokens.map(token => {
    switch (token.type) {
      case 'keyword':
        return `<span class="sql-keyword">${escapeHtml(token.text)}</span>`;
      case 'string':
        return `<span class="sql-string">${escapeHtml(token.text)}</span>`;
      case 'number':
        return `<span class="sql-number">${escapeHtml(token.text)}</span>`;
      case 'comment':
        return `<span class="sql-comment">${escapeHtml(token.text)}</span>`;
      default:
        return escapeHtml(token.text);
    }
  }).join('');
}

function escapeHtml(text: string): string {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}
