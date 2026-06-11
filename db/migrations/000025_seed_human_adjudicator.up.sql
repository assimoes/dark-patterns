  INSERT INTO annotators (kind, model_id, label)
  VALUES ('human', NULL, 'author')
  ON CONFLICT (label) DO NOTHING;