#!/usr/bin/env python3
"""Closed retail ASM grammar and DB byte semantics, including negative controls."""
import unittest
import matrix_operand_schema as s
class SourceExpressions(unittest.TestCase):
 def test_jingle_fallback_counter_is_read(self):
  # FANTASIE _SETDECCOR copies [BX+2]; WAITJINGLE2 decrements it
  # under INT66 function21's no-sound bit. It is not a filler word.
  self.assertEqual(s.numeric_values('_SETDECCOR',['70'],{}),{0:70})
  with self.assertRaises(s.OperandError):s.numeric_values('_SETDECCOR',['?'],{})
 def test_arithmetic(self):
  for expr,want in [('2*60',120),('10*2*15',300),('10*2*5',100),('-(7/2)',-3),('-7/2',-3),('0FFh+1',256),('(SW*4)/4+16',352)]:
   self.assertEqual(s.resolve_expression(expr,{'SW':336}),want)
  for expr in ['2**3','2//3','1<<2','f(3)','__import__("os")','unknown','1/0','(1+2','1 2']:
   with self.assertRaises(s.OperandError):s.resolve_expression(expr,{})
 def test_db(self):
  self.assertEqual(s.db_bytes("'7'+2,' ','7'+1,6+'7','0'+7,'*'-'A'+10",{}),[57,32,56,61,55,243])
  self.assertEqual(len(s.db_bytes('16*16 DUP(0)',{})),256)
  self.assertEqual(len(s.db_bytes('576*2 DUP(0)',{})),1152)
  self.assertEqual(s.db_bytes("2 DUP('A',1+2),-1,256",{}),[65,3,65,3,255,0])
  for expr in ['bad DUP(0)',"'x",'2 DUP(unknown)']:
   with self.assertRaises(s.OperandError):s.db_bytes(expr,{})
if __name__=='__main__':unittest.main()
