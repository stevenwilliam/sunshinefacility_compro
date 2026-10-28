UPDATE parameters SET value_en = '+62 21 1234 5678', value_id = '+62 21 1234 5678', is_placeholder = true, updated_at = now()
 WHERE grp = 'contact' AND key = 'phone' AND value_en = '';
