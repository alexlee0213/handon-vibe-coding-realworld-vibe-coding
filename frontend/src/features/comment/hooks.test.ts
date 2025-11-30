import { describe, it, expect } from 'vitest';
import { commentKeys } from './hooks';

describe('commentKeys', () => {
  describe('all', () => {
    it('returns base key', () => {
      expect(commentKeys.all).toEqual(['comments']);
    });
  });

  describe('lists', () => {
    it('returns list keys', () => {
      expect(commentKeys.lists()).toEqual(['comments', 'list']);
    });
  });

  describe('list', () => {
    it('returns list key with slug', () => {
      expect(commentKeys.list('test-article')).toEqual(['comments', 'list', 'test-article']);
    });

    it('returns different keys for different slugs', () => {
      const key1 = commentKeys.list('article-1');
      const key2 = commentKeys.list('article-2');
      expect(key1).not.toEqual(key2);
    });
  });
});
