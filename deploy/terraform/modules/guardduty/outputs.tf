output "detector_id" {
  value = aws_guardduty_detector.this.id
}

output "sns_topic_arn" {
  value = aws_sns_topic.alerts.arn
}
