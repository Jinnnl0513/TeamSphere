const usernameRegex = /^\w{3,32}$/;
const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export const normalizeText = (value: string) => value.trim();

export const validateUsername = (value: string) => {
  if (!value) return '用户名不能为空。';
  if (!usernameRegex.test(value)) return '用户名必须是 3-32 个字符，只能包含字母、数字和下划线。';
  return '';
};

export const validateEmail = (value: string) => {
  if (!value) return '邮箱不能为空。';
  if (!emailRegex.test(value)) return '邮箱格式不正确。';
  return '';
};

export const validatePassword = (value: string, min = 6, max = 128) => {
  if (!value) return '密码不能为空。';
  if (value.length < min || value.length > max) return `密码长度必须为 ${min} 到 ${max} 个字符。`;
  return '';
};
