def not_matcher(marg):
  def matcher(*args, **kwargs):
    return not marg(*args, **kwargs)
  end
  return matcher
end
